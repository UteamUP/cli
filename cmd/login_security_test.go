package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestLoginRejectsCredentialValuesFromArgv(t *testing.T) {
	originalKey, originalSecret := loginAPIKey, loginAPISecret
	t.Cleanup(func() {
		loginAPIKey, loginAPISecret = originalKey, originalSecret
	})

	for _, testCase := range []struct {
		name   string
		key    string
		secret string
	}{
		{name: "api key", key: "key-value"},
		{name: "api secret", secret: "secret-value"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			loginAPIKey, loginAPISecret = testCase.key, testCase.secret
			err := runLogin(&cobra.Command{}, nil)
			if err == nil || !strings.Contains(err.Error(), "credentials in argv are unsafe") {
				t.Fatalf("expected argv credential rejection, got %v", err)
			}
		})
	}
}
