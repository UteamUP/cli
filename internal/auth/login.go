package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// PromptCredentials interactively prompts for email and password.
func PromptCredentials() (email, password string, err error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Email: ")
	email, err = reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("reading email: %w", err)
	}
	email = strings.TrimSpace(email)

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", "", fmt.Errorf("reading password: %w", err)
	}
	fmt.Println() // newline after hidden input
	password = string(passwordBytes)

	if email == "" || password == "" {
		return "", "", fmt.Errorf("email and password are required")
	}

	return email, password, nil
}

// PromptAPIKey interactively prompts for API key and secret.
func PromptAPIKey() (apiKey, secret string, err error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("API Key (32 characters): ")
	apiKey, err = reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("reading API key: %w", err)
	}
	apiKey = strings.TrimSpace(apiKey)

	fmt.Print("Secret (64+ characters): ")
	secretBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", "", fmt.Errorf("reading secret: %w", err)
	}
	fmt.Println()
	secret = string(secretBytes)

	return apiKey, secret, nil
}
