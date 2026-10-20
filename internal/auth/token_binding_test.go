package auth

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

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

func TestTokenPersistenceUsesPrivatePermissionsAndRejectsPublicCache(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if err := SaveToken(&TokenData{
		APIOrigin:   "https://api.uteamup.com",
		AccessToken: "test-token",
		ExpiresAt:   time.Now().Add(time.Hour),
		Profile:     "production",
	}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".uteamup", tokenFileName)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("token permissions = %04o, want 0600", info.Mode().Perm())
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(path, 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadToken(); err == nil {
			t.Fatal("public token cache accepted")
		}
	}
}

func TestLoadTokenWithoutConfigDirectoryIsUnauthenticated(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	token, err := LoadToken()
	if err != nil || token != nil {
		t.Fatalf("LoadToken = (%v, %v), want (nil, nil)", token, err)
	}
}
