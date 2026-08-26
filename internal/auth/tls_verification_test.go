package auth

import (
	"crypto/tls"
	"net/http"
	"testing"
)

// FetchTenantInfo and FetchAllTenants hardcoded InsecureSkipVerify: true, ignoring the
// --insecure flag entirely. Every `uteamup tenant` call against production therefore
// accepted any certificate, so anyone on the network path could present their own and
// read the bearer token out of the Authorization header.
//
// A working --insecure flag already existed and was plumbed through the API client; these
// two functions simply bypassed it.

func TestSkipTLSVerifyIsOffByDefault(t *testing.T) {
	hosts := []string{
		"https://api.uteamup.com",
		"https://devback.uteamup.com",
		"https://localhost:5002",
		"https://127.0.0.1:5002",
	}

	for _, h := range hosts {
		if skipTLSVerifyFor(h, false) {
			t.Errorf("skipTLSVerifyFor(%q, insecure=false) = true, want false: "+
				"certificate verification must never be skipped unless explicitly requested", h)
		}
	}
}

func TestSkipTLSVerifyRejectedForRemoteHostsEvenWhenRequested(t *testing.T) {
	// The case that mattered: --insecure against production must not disable verification.
	remote := []string{
		"https://api.uteamup.com",
		"https://devback.uteamup.com",
		"https://prufaback.uteamup.com",
		"https://uteamup.com",
		"https://10.0.0.5",        // private, but still a real network path
		"https://169.254.169.254", // cloud metadata
	}

	for _, h := range remote {
		if skipTLSVerifyFor(h, true) {
			t.Errorf("skipTLSVerifyFor(%q, insecure=true) = true, want false: "+
				"a remote host has a network path to intercept, so the bearer token would be exposed", h)
		}
	}
}

func TestSkipTLSVerifyAllowedForLoopbackWhenRequested(t *testing.T) {
	// Local development against a self-signed certificate stays workable.
	loopback := []string{
		"https://localhost:5002",
		"https://localhost:3000/",
		"https://127.0.0.1:5002",
		"https://[::1]:5002",
	}

	for _, h := range loopback {
		if !skipTLSVerifyFor(h, true) {
			t.Errorf("skipTLSVerifyFor(%q, insecure=true) = false, want true: "+
				"loopback development against a self-signed certificate must remain possible", h)
		}
	}
}

func TestSkipTLSVerifyHandlesUnparseableBaseURL(t *testing.T) {
	// Fail closed: an address we cannot understand is not one we can call loopback.
	if skipTLSVerifyFor("://not a url", true) {
		t.Error("an unparseable base URL must not enable insecure TLS")
	}
}

func TestTenantHTTPClientEnforcesMinimumTLSVersion(t *testing.T) {
	client := tenantHTTPClient("https://api.uteamup.com", false)

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport = %T, want *http.Transport", client.Transport)
	}

	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %x, want TLS 1.2 (%x)",
			transport.TLSClientConfig.MinVersion, tls.VersionTLS12)
	}

	if transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify must be false for a remote host")
	}
}
