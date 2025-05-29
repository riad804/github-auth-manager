package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// PromptForInput prompts the user for input with a given message.
func PromptForInput(promptMessage string, sensitive bool) (string, error) {
	fmt.Print(promptMessage)
	if sensitive {
		bytePassword, err := term.ReadPassword(int(syscall.Stdin)) // On Windows, os.Stdin may not be a file descriptor
		// bytePassword, err := term.ReadPassword(int(os.Stdin.Fd())) // Use this for Unix-like
		if err != nil {
			return "", fmt.Errorf("failed to read sensitive input: %w", err)
		}
		fmt.Println() // Add a newline after password input
		return strings.TrimSpace(string(bytePassword)), nil
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	return strings.TrimSpace(input), nil
}
