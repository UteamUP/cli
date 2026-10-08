package auth

import (
	"bufio"
	"context"
	"io"
	"os"

	"golang.org/x/term"
)

type samlTerminalIO struct {
	io.Reader
	io.Writer
}

// Set terminal mode before starting the read so cancellation always restores it.
func promptSamlSecret(ctx context.Context, reader *bufio.Reader, prompt string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return awaitSamlInput(ctx, func() (string, error) { return PromptSecret(reader, prompt) })
	}
	state, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, state)
	console := term.NewTerminal(samlTerminalIO{Reader: reader, Writer: os.Stderr}, "")
	return awaitSamlInput(ctx, func() (string, error) { return console.ReadPassword(prompt) })
}

func awaitSamlInput(ctx context.Context, read func() (string, error)) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	type input struct {
		code string
		err  error
	}
	result := make(chan input, 1)
	go func() {
		code, err := read()
		result <- input{code: code, err: err}
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case value := <-result:
		if err := ctx.Err(); err != nil {
			return "", err
		}
		return value.code, value.err
	}
}
