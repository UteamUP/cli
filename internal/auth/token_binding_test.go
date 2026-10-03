package auth

import "testing"

func TestTokenBindingRejectsLegacyChangedOriginAndProfile(t *testing.T) {
	token := &TokenData{APIOrigin: "https://example.com", Profile: "production"}
	if err := token.ValidateBinding("https://EXAMPLE.com:443/", "production"); err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]string{{"https://other.example", "production"}, {"https://example.com", "other"}, {"http://example.com", "production"}} {
		if token.ValidateBinding(pair[0], pair[1]) == nil {
			t.Fatal(pair)
		}
	}
	token.APIOrigin = ""
	if token.ValidateBinding("https://example.com", "production") == nil {
		t.Fatal("legacy cache accepted")
	}
}
