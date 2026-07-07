package cmd

import "testing"

func TestEnvironmentFromBaseURL(t *testing.T) {
	cases := []struct {
		baseURL  string
		expected string
	}{
		{"https://api.uteamup.com", "production"},
		{"https://api.uteamup.com/", "production"},
		{"https://devback.uteamup.com", "dev"},
		{"https://DEVBACK.uteamup.com", "dev"},
		{"https://prufaback.uteamup.com", "staging"},
		{"https://staging.uteamup.com", "staging"},
		{"https://localhost:5002", "localhost"},
		{"https://127.0.0.1:5002", "localhost"},
		{"https://something-else.com", "production"},
	}

	for _, c := range cases {
		t.Run(c.baseURL, func(t *testing.T) {
			got := environmentFromBaseURL(c.baseURL)
			if got != c.expected {
				t.Errorf("environmentFromBaseURL(%q) = %q, want %q", c.baseURL, got, c.expected)
			}
		})
	}
}
