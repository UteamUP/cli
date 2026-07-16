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
