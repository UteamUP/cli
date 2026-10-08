package auth

import (
	"context"
	"errors"
	"testing"
)

func TestSamlProofInputCanCancelWithoutWaitingForInput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started, release, finished := make(chan struct{}), make(chan struct{}), make(chan struct{})
	defer close(release)
	go func() {
		defer close(finished)
		code, err := awaitSamlInput(ctx, func() (string, error) { close(started); <-release; return "PRIVATE-CODE", nil })
		if code != "" || !errors.Is(err, context.Canceled) {
			t.Errorf("cancelled proof input was accepted: %v", err)
		}
	}()
	<-started
	cancel()
	<-finished
}

func TestSamlProofInputReturnsTheTypedCodeOrReadFailure(t *testing.T) {
	code, err := awaitSamlInput(context.Background(), func() (string, error) { return "123456", nil })
	if err != nil || code != "123456" {
		t.Fatal("proof input was lost")
	}
	want := errors.New("input closed")
	_, err = awaitSamlInput(context.Background(), func() (string, error) { return "", want })
	if !errors.Is(err, want) {
		t.Fatal("input failure was lost")
	}
}
