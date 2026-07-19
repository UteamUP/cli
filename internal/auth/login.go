package auth

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// readSecretPlain reads one line as a secret without hiding it. It reuses the
// caller's reader because bufio may already hold bytes past the previous line;
// reading os.Stdin directly here would discard them.
func readSecretPlain(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil && !(errors.Is(err, io.EOF) && line != "") {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// readSecret prompts for and reads a secret, hiding the input when stdin is a
// real console.
//
// When stdin is not a console, term.ReadPassword cannot switch the handle to
// raw mode and fails — on Windows with the opaque "The handle is invalid".
// That is the normal case under Git Bash/MinTTY (whose pty is a named pipe,
// not a console), under pipes, and in CI, so fall back to a visible read
// instead of failing the login outright.
func readSecret(reader *bufio.Reader, prompt string) (string, error) {
	fd := int(os.Stdin.Fd())

	if term.IsTerminal(fd) {
		fmt.Print(prompt)
		secretBytes, err := term.ReadPassword(fd)
		if err != nil {
			return "", err
		}
		fmt.Println() // newline after hidden input
		return string(secretBytes), nil
	}

	fmt.Fprintln(os.Stderr,
		"warning: stdin is not a console (Git Bash/MinTTY, pipe, or CI) — input will NOT be hidden")
	fmt.Print(prompt)
	return readSecretPlain(reader)
}

// PromptCredentials interactively prompts for email and password.
func PromptCredentials() (email, password string, err error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Email: ")
	email, err = reader.ReadString('\n')
	if err != nil && !(errors.Is(err, io.EOF) && email != "") {
		return "", "", fmt.Errorf("reading email: %w", err)
	}
	email = strings.TrimSpace(email)

	password, err = readSecret(reader, "Password: ")
	if err != nil {
		return "", "", fmt.Errorf("reading password: %w", err)
	}

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
	if err != nil && !(errors.Is(err, io.EOF) && apiKey != "") {
		return "", "", fmt.Errorf("reading API key: %w", err)
	}
	apiKey = strings.TrimSpace(apiKey)

	secret, err = readSecret(reader, "Secret (64+ characters): ")
	if err != nil {
		return "", "", fmt.Errorf("reading secret: %w", err)
	}

	return apiKey, secret, nil
}
