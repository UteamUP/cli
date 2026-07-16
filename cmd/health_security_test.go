package cmd

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestHealthClientRequiresTLS12OrNewer(t *testing.T) {
	client := newHealthClient()
	transport, ok := client.Transport.(*http.Transport)
	if !ok || transport.TLSClientConfig == nil {
		t.Fatal("health client must use an explicit TLS transport")
	}
	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Fatalf("minimum TLS version = %d, want TLS 1.2", transport.TLSClientConfig.MinVersion)
	}
}
