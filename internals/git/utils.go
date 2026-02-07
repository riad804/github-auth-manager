package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// FindGitRepo finds .git dir upwards
func FindGitRepo() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not in a Git repo")
		}
		dir = parent
	}
}

// GetRemoteURL gets git remote url
func GetRemoteURL(remote string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", remote)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// DetectProvider from URL
func DetectProvider(url string) string {
	if strings.Contains(url, "github.com") {
		return "github"
	} else if strings.Contains(url, "gitlab.com") {
		return "gitlab"
	} else if strings.Contains(url, "bitbucket.org") {
		return "bitbucket"
	}
	return "unknown"
}

// GetGlobalConfig gets git config
func GetGlobalConfig(key string) (string, error) {
	cmd := exec.Command("git", "config", "--global", key)
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

// SetGlobalConfigAdd adds to multi-value config
func SetGlobalConfigAdd(key, value string) error {
	cmd := exec.Command("git", "config", "--global", "--add", key, value)
	return cmd.Run()
}

// DetectOS simple
func DetectOS() string {
	return runtime.GOOS
}
