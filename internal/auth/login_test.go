package auth

import (
	"bufio"
	"strings"
	"testing"
)

func TestReadSecretPlainReadsLine(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("hunter2\n"))

	got, err := readSecretPlain(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hunter2" {
		t.Errorf("got %q, want %q", got, "hunter2")
	}
}

// A piped secret often arrives without a trailing newline; ReadString reports
// io.EOF alongside the data, which must not be treated as a failure.
func TestReadSecretPlainAcceptsMissingTrailingNewline(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("hunter2"))

	got, err := readSecretPlain(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hunter2" {
		t.Errorf("got %q, want %q", got, "hunter2")
	}
}

// Windows line endings must not survive into the password sent to the backend,
// or login fails with an opaque 401.
func TestReadSecretPlainStripsCarriageReturn(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("hunter2\r\n"))

	got, err := readSecretPlain(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hunter2" {
		t.Errorf("got %q, want %q", got, "hunter2")
	}
}

func TestReadSecretPlainPreservesInternalSpaces(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("correct horse battery staple\n"))

	got, err := readSecretPlain(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "correct horse battery staple" {
		t.Errorf("got %q, want the spaces preserved", got)
	}
}

func TestReadSecretPlainErrorsOnEmptyInput(t *testing.T) {
	r := bufio.NewReader(strings.NewReader(""))

	if _, err := readSecretPlain(r); err == nil {
		t.Error("expected an error when stdin is empty, got nil")
	}
}

// PromptCredentials reads the email itself and then hands the same reader to
// readSecret. If the secret read used a fresh reader it would drop whatever
// bufio had already buffered, so the password must come back intact here.
func TestReadSecretPlainContinuesFromSharedReader(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("user@example.com\nhunter2\n"))

	email, err := r.ReadString('\n')
	if err != nil {
		t.Fatalf("unexpected error reading email: %v", err)
	}
	if strings.TrimSpace(email) != "user@example.com" {
		t.Fatalf("got email %q", strings.TrimSpace(email))
	}

	got, err := readSecretPlain(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hunter2" {
		t.Errorf("got %q, want %q — buffered bytes were lost", got, "hunter2")
	}
}
