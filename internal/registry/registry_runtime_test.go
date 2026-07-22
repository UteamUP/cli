package registry

import (
	"errors"
	"testing"

	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func TestBuildCommandsCreatesAPIClientWhenActionExecutes(t *testing.T) {
	wantErr := errors.New("client factory reached")
	called := 0
	factory := func() (*client.APIClient, error) {
		called++
		return nil, wantErr
	}

	registry := &Registry{domains: []*Domain{{
		Name: "sample",
		Actions: []Action{{
			Name:     "list",
			ToolName: "SampleList",
		}},
	}}}
	format := "json"
	commands := registry.BuildCommands(
		factory,
		logging.New(logging.LevelError),
		&format,
		&ExportConfig{},
	)

	if called != 0 {
		t.Fatalf("client factory called during command registration: %d", called)
	}

	commands[0].SetArgs([]string{"list"})
	err := commands[0].Execute()
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want %v", err, wantErr)
	}
	if called != 1 {
		t.Fatalf("client factory calls = %d, want 1", called)
	}
}
