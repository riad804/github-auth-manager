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

	"github.com/go-git/go-git/v5"
	gc "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// var activeContext *Context
// var token string

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
	command := gitArgs[0]
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

	if isInsideRepo && activeContext != nil && token != "" {
		switch command {
		case "pull":
			return handlePullWithGoGit(repoRoot, gitArgs, activeContext, token, outW, errW)
		case "push":
			return handlePushWithGoGit(repoRoot, gitArgs, activeContext, token, outW, errW)
		case "fetch":
			return handleFetchWithGoGit(repoRoot, gitArgs, activeContext, token, outW, errW)
		}
	}

	if contextName != "" {
		fmt.Fprintf(errW, "[GHAM] Using token from context '%s' for GitHub operations.\n", contextName)
	}

	// For other commands (including clone), use the original exec-based approach
	return executeWithOSCommand(gitArgs, cwd, repoRoot, isInsideRepo, outW, errW, activeContext, token)
}

// executeWithOSCommand handles non go-git commands using OS exec
func executeWithOSCommand(gitArgs []string, cwd, repoRoot string, isInsideRepo bool, outW, errW io.Writer, activeContext *config.Context, token string) error {
	cmdArgs := []string{}
	envVars := os.Environ()

	if activeContext != nil && token != "" {
		// Disable credential helpers
		cmdArgs = append(cmdArgs, "-c", "credential.helper=")

		// Set user config if provided
		if activeContext.Username != "" && activeContext.Username != config.DefaultUserName {
			cmdArgs = append(cmdArgs, "-c", fmt.Sprintf("user.name=%s", activeContext.Username))
		}
		if activeContext.Email != "" {
			cmdArgs = append(cmdArgs, "-c", fmt.Sprintf("user.email=%s", activeContext.Email))
		}

		// Handle clone command separately
		if len(gitArgs) > 1 && gitArgs[0] == "clone" {
			origURL := gitArgs[1]
			// Convert SSH URLs to HTTPS
			if strings.HasPrefix(origURL, "git@") {
				origURL = convertSSHtoHTTPS(origURL)
			}
			if strings.HasPrefix(origURL, "https://") {
				u, err := url.Parse(origURL)
				if err != nil {
					fmt.Fprintf(errW, "Warning: failed to parse URL %s: %v\n", origURL, err)
				} else {
					u.User = url.UserPassword(activeContext.Username, token)
					gitArgs[1] = u.String()
					fmt.Fprintf(errW, "[GHAM] Using authenticated URL: %s\n", u.Redacted())
				}
			}
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

	return gitCommand.Run()
}

func handlePullWithGoGit(repoRoot string, gitArgs []string, ctx *config.Context, token string, outW, errW io.Writer) error {
	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	auth := &http.BasicAuth{
		Username: ctx.Username,
		Password: token,
	}

	opts := &git.PullOptions{
		RemoteName: "origin",
		Auth:       auth,
		Progress:   outW,
	}

	// Handle branch specification if provided
	if len(gitArgs) > 1 {
		for i, arg := range gitArgs[1:] {
			if arg == "-b" || arg == "--branch" {
				if i+1 < len(gitArgs[1:]) {
					branch := gitArgs[i+2]
					opts.ReferenceName = plumbing.ReferenceName("refs/heads/" + branch)
				}
			}
		}
	}

	fmt.Fprintf(errW, "[GHAM] Pulling with token from context '%s'\n", ctx.Name)
	return w.Pull(opts)
}

func handlePushWithGoGit(repoRoot string, gitArgs []string, ctx *config.Context, token string, outW, errW io.Writer) error {
	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	auth := &http.BasicAuth{
		Username: ctx.Username,
		Password: token,
	}

	opts := &git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
		Progress:   outW,
	}

	// Handle branch specification
	if len(gitArgs) > 1 {
		refSpecs := []gc.RefSpec{}
		for _, arg := range gitArgs[1:] {
			if !strings.HasPrefix(arg, "-") && arg != "origin" {
				refSpecs = append(refSpecs, gc.RefSpec("refs/heads/"+arg+":refs/heads/"+arg))
			}
		}
		if len(refSpecs) > 0 {
			opts.RefSpecs = refSpecs
		}
	}

	fmt.Fprintf(errW, "[GHAM] Pushing with token from context '%s'\n", ctx.Name)
	return repo.Push(opts)
}

func handleFetchWithGoGit(repoRoot string, gitArgs []string, ctx *config.Context, token string, outW, errW io.Writer) error {
	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	auth := &http.BasicAuth{
		Username: ctx.Username,
		Password: token,
	}

	opts := &git.FetchOptions{
		RemoteName: "origin",
		Auth:       auth,
		Progress:   outW,
	}

	fmt.Fprintf(errW, "[GHAM] Fetching with token from context '%s'\n", ctx.Name)
	return repo.Fetch(opts)
}

// Helper functions
func convertSSHtoHTTPS(sshURL string) string {
	return strings.NewReplacer(
		"git@github.com:", "https://github.com/",
		".git", "",
	).Replace(sshURL) + ".git"
}
