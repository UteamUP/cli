package cmd

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func healthTransport(t *testing.T, baseURL string) *http.Transport {
	t.Helper()
	client := newHealthClient(baseURL)
	transport, ok := client.Transport.(*http.Transport)
	if !ok || transport.TLSClientConfig == nil {
		t.Fatal("health client must use an explicit TLS transport")
	}
	return transport
}

func TestHealthClientRequiresTLS12OrNewer(t *testing.T) {
	transport := healthTransport(t, "https://api.uteamup.com")
	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Fatalf("minimum TLS version = %d, want TLS 1.2", transport.TLSClientConfig.MinVersion)
	}
}

// The health probe reads the same --insecure flag as every other client, and it used to
// pass it straight into TLSClientConfig. That let `uteamup health --insecure` accept any
// certificate for a remote host, so an on-path attacker could impersonate the API
// (UTP-CLI-INSECURE-TLS). Verification may only be skipped for loopback.
func TestHealthClientNeverSkipsVerifyForRemoteHosts(t *testing.T) {
	original := insecure
	t.Cleanup(func() { insecure = original })
	insecure = true

	remotes := []string{
		"https://api.uteamup.com",
		"https://devback.uteamup.com",
		"https://prufaback.uteamup.com",
		"https://uteamup.com.evil.example",
	}
	for _, baseURL := range remotes {
		if healthTransport(t, baseURL).TLSClientConfig.InsecureSkipVerify {
			t.Errorf("InsecureSkipVerify must be false for remote host %s even with --insecure", baseURL)
		}
	}
}

func TestHealthClientHonoursInsecureForLoopbackOnly(t *testing.T) {
	original := insecure
	t.Cleanup(func() { insecure = original })

	insecure = true
	for _, baseURL := range []string{"https://localhost:5002", "https://127.0.0.1:5002"} {
		if !healthTransport(t, baseURL).TLSClientConfig.InsecureSkipVerify {
			t.Errorf("--insecure must still be honoured for loopback host %s", baseURL)
		}
	}

	insecure = false
	if healthTransport(t, "https://localhost:5002").TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify must be false for loopback when --insecure was not passed")
	}
}
