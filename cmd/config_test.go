package cmd

import "testing"

func TestSafeConfigDisplayValueRedactsCredentials(t *testing.T) {
	for _, key := range []string{"apiKey", "apikey", "secret"} {
		if got := safeConfigDisplayValue(key, "must-not-leak"); got != "***" {
			t.Fatalf("%s display value = %q", key, got)
		}
	}
	if got := safeConfigDisplayValue("logLevel", "DEBUG"); got != "DEBUG" {
		t.Fatalf("non-secret display value = %q", got)
	}
}
