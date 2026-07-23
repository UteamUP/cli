package client

import "testing"

func TestValidateBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "production", value: "https://api.uteamup.com"},
		{name: "localhost TLS", value: "https://localhost:5002"},
		{name: "http rejected", value: "http://api.uteamup.com", wantErr: true},
		{name: "credentials rejected", value: "https://user:pass@api.uteamup.com", wantErr: true},
		{name: "path rejected", value: "https://api.uteamup.com/redirect", wantErr: true},
		{name: "query rejected", value: "https://api.uteamup.com?target=elsewhere", wantErr: true},
		{name: "relative rejected", value: "//api.uteamup.com", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateBaseURL(test.value)
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateBaseURL(%q) error = %v, wantErr %t", test.value, err, test.wantErr)
			}
		})
	}
}

func TestAppendQueryString(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		query string
		want  string
	}{
		{name: "first query", url: "https://api.uteamup.com/api/workorder/search", query: "page=1", want: "https://api.uteamup.com/api/workorder/search?page=1"},
		{name: "existing query", url: "https://api.uteamup.com/api/workorder/search?query=pump", query: "page=1&pageSize=25", want: "https://api.uteamup.com/api/workorder/search?query=pump&page=1&pageSize=25"},
		{name: "empty query", url: "https://api.uteamup.com/api/workorder/search?query=pump", want: "https://api.uteamup.com/api/workorder/search?query=pump"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := appendQueryString(test.url, test.query); got != test.want {
				t.Fatalf("appendQueryString(%q, %q) = %q, want %q", test.url, test.query, got, test.want)
			}
		})
	}
}
