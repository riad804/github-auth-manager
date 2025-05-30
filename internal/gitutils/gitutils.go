package gitutils

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/riad804/github-auth-manager/internal/config"
	"github.com/riad804/github-auth-manager/internal/keyring"
)

// FindRepoRoot traverses up from the given path to find a .git directory.
// Returns the absolute path to the directory containing .git, or an error if not found.
func FindRepoRoot(startPath string) (string, error) {
	currentPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for '%s': %w", startPath, err)
	}
	fi, err := os.Stat(currentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("path '%s' does not exist", startPath)
		}
		return "", fmt.Errorf("failed to stat path '%s': %w", startPath, err)
	}
	if !fi.IsDir() {
		currentPath = filepath.Dir(currentPath)
	}
	originalPathForError := currentPath

	for {
		gitDirPath := filepath.Join(currentPath, ".git")
		stat, err := os.Stat(gitDirPath)
		if err == nil && stat.IsDir() {
			return currentPath, nil
		}
		if err == nil && !stat.IsDir() {
			return currentPath, nil
		}
		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			break
		}
		currentPath = parent
	}
	return "", fmt.Errorf("not a git repository (or any of the parent directories of '%s')", originalPathForError)
}

// getHostFromURL parses a git remote URL and returns the host.
// e.g., https://github.com/user/repo.git -> github.com
//
//	git@github.com:user/repo.git -> github.com
func getHostFromURL(remoteURL string) (string, error) {
	if strings.Contains(remoteURL, "@") && strings.Contains(remoteURL, ":") &&
		!strings.HasPrefix(remoteURL, "http") && !strings.HasPrefix(remoteURL, "ssh://") {
		parts := strings.SplitN(remoteURL, "@", 2)
		if len(parts) < 2 {
			return "", fmt.Errorf("malformed SCP-like URL: %s", remoteURL)
		}
		hostAndPath := strings.SplitN(parts[1], ":", 2)
		if len(hostAndPath) < 1 {
			return "", fmt.Errorf("malformed SCP-like URL (no host): %s", remoteURL)
		}
		return hostAndPath[0], nil
	}
	parsedURL, err := url.Parse(remoteURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse remote URL '%s': %w", remoteURL, err)
	}
	host := parsedURL.Hostname()
	if host == "" {
		return "", fmt.Errorf("could not determine host from URL '%s'", remoteURL)
	}
	return host, nil
}

// ExecuteGitCommandWithContext wraps a git command, injecting context-specific credentials.
// Takes io.Writer for stdout and stderr for better testability and control.
func ExecuteGitCommandWithContext(gitArgs []string, outW, errW io.Writer) error {
	if len(gitArgs) == 0 {
		return fmt.Errorf("no git command provided")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	var activeContext *config.Context
	var token string
	var contextName string

	repoRoot, err := FindRepoRoot(cwd)
	isInsideRepo := err == nil

	if isInsideRepo {
		repoCtxName, found := config.GetRepoContextName(repoRoot)
		if found {
			ctx, ctxFound := config.FindContext(repoCtxName)
			if ctxFound {
				activeContext = ctx
				contextName = ctx.Name
				retrievedToken, tokenErr := keyring.GetToken(ctx.Name)
				if tokenErr != nil {
					fmt.Fprintf(errW, "Warning: GHAM context '%s' is active but token could not be retrieved: %v\n", ctx.Name, tokenErr)
					fmt.Fprintln(errW, "Git command will proceed without GHAM token injection.")
				} else {
					token = retrievedToken
				}
			} else {
				fmt.Fprintf(errW, "Warning: Context '%s' assigned to repo but not found. Using system Git config.\n", repoCtxName)
			}
		}
	}

	cmdArgs := make([]string, 0, len(gitArgs)+4) // Pre-allocate with some extra capacity
	envVars := os.Environ()

	if activeContext != nil && token != "" {
		if activeContext.Username != "" && activeContext.Username != config.DefaultUserName {
			cmdArgs = append(cmdArgs, "-c", fmt.Sprintf("user.name=%s", activeContext.Username))
		}
		if activeContext.Email != "" {
			cmdArgs = append(cmdArgs, "-c", fmt.Sprintf("user.email=%s", activeContext.Email))
		}

		command := gitArgs[0]
		switch command {
		case "clone":
			if len(gitArgs) > 1 {
				origURL := gitArgs[1]
				if strings.HasPrefix(origURL, "https://") {
					u, err := url.Parse(origURL)
					if err == nil {
						u.User = url.UserPassword(activeContext.Username, token)
						gitArgs[1] = u.String()
					}
				}
			}
		case "pull", "push", "fetch":
			if isInsideRepo {
				originURLCmd := exec.Command("git", "config", "--get", "remote.origin.url")
				originURLCmd.Dir = repoRoot
				output, err := originURLCmd.Output()
				if err == nil {
					remoteOriginURL := strings.TrimSpace(string(output))
					if strings.HasPrefix(remoteOriginURL, "https://") {
						u, err := url.Parse(remoteOriginURL)
						if err == nil {
							u.User = url.UserPassword(activeContext.Username, token)
							tempRemote := u.String()
							cmdArgs = append(cmdArgs, "-c", fmt.Sprintf("remote.origin.url=%s", tempRemote))
						}
					}
				}
			}
		}

		if contextName != "" {
			fmt.Fprintf(errW, "[GHAM] Using token from context '%s' for GitHub operations.\n", contextName)
		}
	}

	cmdArgs = append(cmdArgs, gitArgs...)

	gitCommand := exec.Command("git", cmdArgs...)
	gitCommand.Env = envVars
	gitCommand.Stdin = os.Stdin
	gitCommand.Stdout = outW
	gitCommand.Stderr = errW

	if isInsideRepo {
		gitCommand.Dir = repoRoot
	} else {
		gitCommand.Dir = cwd
	}

	err = gitCommand.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("git command failed with exit code %d: %s", exitErr.ExitCode(), exitErr.Stderr)
		}
		return fmt.Errorf("failed to execute git command: %w", err)
	}

	return nil
}
