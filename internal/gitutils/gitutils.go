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

	// Ensure the start path itself exists and is a directory if we're looking from it
	// This check is more relevant if startPath is user input for a directory,
	// os.Getwd() will generally be valid.
	fi, err := os.Stat(currentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("path '%s' does not exist", startPath)
		}
		return "", fmt.Errorf("failed to stat path '%s': %w", startPath, err)
	}
	if !fi.IsDir() { // If startPath is a file, start from its directory
		currentPath = filepath.Dir(currentPath)
	}

	originalPathForError := currentPath // Save for a more user-friendly error message

	for {
		gitDirPath := filepath.Join(currentPath, ".git")
		stat, err := os.Stat(gitDirPath)
		if err == nil && stat.IsDir() {
			return currentPath, nil // Found .git directory
		}
		// Also check for .git file (worktrees)
		if err == nil && !stat.IsDir() {
			// If .git is a file, it might be a worktree. Read its content.
			// Content is usually: gitdir: /path/to/.git/worktrees/worktree-name
			// For simplicity, we'll consider the directory containing this .git file as the root.
			// A more robust solution would parse this file.
			return currentPath, nil
		}

		parent := filepath.Dir(currentPath)
		if parent == currentPath { // Reached filesystem root
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
	// Handle SCP-like syntax first: git@host:path
	if strings.Contains(remoteURL, "@") && strings.Contains(remoteURL, ":") && !strings.HasPrefix(remoteURL, "http") && !strings.HasPrefix(remoteURL, "ssh://") {
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

	// Handle standard URLs
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
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	var activeContext *config.Context
	var token string
	var contextName string // To store the name of the active context for messages

	repoRoot, err := FindRepoRoot(cwd)
	isInsideRepo := err == nil

	if isInsideRepo {
		repoCtxName, found := config.GetRepoContextName(repoRoot)
		if found {
			ctx, ctxFound := config.FindContext(repoCtxName)
			if ctxFound {
				activeContext = ctx
				contextName = ctx.Name // Store for messages
				retrievedToken, tokenErr := keyring.GetToken(ctx.Name)
				if tokenErr != nil {
					// Don't fail the command immediately, but warn. Some git commands don't need auth.
					fmt.Fprintf(errW, "Warning: GHAM context '%s' is active for this repository, but could not retrieve its token: %v\n", ctx.Name, tokenErr)
					fmt.Fprintln(errW, "Git command will proceed without GHAM token injection.")
				} else {
					token = retrievedToken
				}
			} else {
				fmt.Fprintf(errW, "Warning: Repository '%s' is assigned to GHAM context '%s', but this context definition was not found. Using system Git config.\n", repoRoot, repoCtxName)
			}
		} else {
			// No GHAM context assigned to this repository.
			// fmt.Fprintln(outW, "No GHAM context assigned to this repository. Using system Git config.") // Can be verbose
		}
	} else {
		// Not inside a known git repository.
		// This could be 'git clone'. For 'git clone', we might want to determine the target host
		// and see if a "default" context for that host exists, or use a globally specified default.
		// For MVP, we'll primarily rely on context if *inside* an assigned repo.
		// For 'git clone https://github.com/...', we can try to apply token if a context could be determined
		// (e.g. a default context, or one matching the host if we implement host-based defaults later).
		// For now, if not in repo, no specific GHAM context is auto-applied unless 'clone' is enhanced.
	}

	// Prepare git command arguments
	cmdArgs := []string{}   // For git itself
	envVars := os.Environ() // Current environment variables

	if activeContext != nil {
		// Set Git user/email for this command execution
		// These override .git/config settings for this command only
		if activeContext.Username != "" && activeContext.Username != config.DefaultUserName {
			cmdArgs = append(cmdArgs, "-c", fmt.Sprintf("user.name=%s", activeContext.Username))
		}
		if activeContext.Email != "" {
			cmdArgs = append(cmdArgs, "-c", fmt.Sprintf("user.email=%s", activeContext.Email))
		}

		if token != "" {
			// Determine the host for PAT injection.
			// This is tricky for commands like `git push` where the remote URL isn't an argument.
			// For `clone`, it's easier.
			// Let's assume for now this applies to github.com or a host matching a future config.
			// For MVP, we'll hardcode for github.com and print a message.
			// More advanced: parse remote URL for 'push'/'pull' if possible, or allow specifying host in context.
			targetHost := "github.com" // Default assumption

			// Attempt to find URL in args (e.g., for clone)
			// A more robust way for push/pull would be to read 'remote.origin.url' from git config if context is active
			for _, arg := range gitArgs {
				if strings.HasPrefix(arg, "https://") || strings.HasPrefix(arg, "git@") {
					host, hostErr := getHostFromURL(arg)
					if hostErr == nil && host != "" {
						targetHost = host
						break
					}
				}
			}
			// If we are inside a repo and have an active context, try to get host from 'origin' remote
			if isInsideRepo && activeContext != nil && (gitArgs[0] == "push" || gitArgs[0] == "pull" || gitArgs[0] == "fetch") {
				originURLCmd := exec.Command("git", "config", "--get", "remote.origin.url")
				originURLCmd.Dir = repoRoot // Run in repo root
				if output, err := originURLCmd.Output(); err == nil {
					remoteOriginURL := strings.TrimSpace(string(output))
					if host, hostErr := getHostFromURL(remoteOriginURL); hostErr == nil && host != "" {
						targetHost = host
					}
				}
			}

			// Only inject PAT for HTTPS URLs
			// For SSH, GCM Core or SSH agent with the correct key handles auth.
			// GHAM's role for SSH would be to manage which SSH key is active (Post-MVP via GIT_SSH_COMMAND).
			// This PAT injection is specific to HTTPS.
			header := fmt.Sprintf("AUTHORIZATION: Bearer %s", token)
			// The key for extraheader needs to be specific to the host.
			// Example: http.https://github.com/.extraheader
			// This setup assumes the token is for `targetHost`.
			cmdArgs = append(cmdArgs, "-c", fmt.Sprintf("http.https://%s/.extraheader=%s", targetHost, header))
			if activeContext != nil {
				fmt.Fprintf(errW, "[GHAM] Using token from context '%s' for Git operations with host '%s'.\n", contextName, targetHost)
			}
		}
	}
	cmdArgs = append(cmdArgs, gitArgs...) // Add the actual git command and its arguments

	// Execute git command
	// fmt.Fprintf(outW, "Executing: git %s\n", strings.Join(cmdArgs, " ")) // Debug: print command
	gitCommand := exec.Command("git", cmdArgs...)
	gitCommand.Env = envVars // Pass environment
	gitCommand.Stdin = os.Stdin
	gitCommand.Stdout = outW
	gitCommand.Stderr = errW
	if isInsideRepo { // Run command from repo root if context is active or just generally if inside a repo
		gitCommand.Dir = repoRoot
	} else {
		gitCommand.Dir = cwd // Run from current working dir if not in a repo (e.g. clone)
	}

	err = gitCommand.Run()
	if err != nil {
		// The error message from git (via exitErr.Stderr) is usually already printed by cmd.Stderr = errW
		// So, just returning the original error is often fine.
		// We can wrap it to indicate it came from the git execution.
		if exitErr, ok := err.(*exec.ExitError); ok {
			// ExitError.Error() already includes "exit status X"
			return fmt.Errorf("git command failed: %s", exitErr.Error())
		}
		return fmt.Errorf("failed to execute git command: %w", err)
	}
	return nil
}
