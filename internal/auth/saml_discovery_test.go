package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/logging"
)

func TestSamlDiscoveryUsesOnlyNormalizedDomainOnTheSelectedBackend(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || !reflect.DeepEqual(body, map[string]string{"domain": "iteggs.com"}) {
			t.Errorf("discovery sent an account or changed the domain: %v", body)
		}
		if r.Method != "POST" || r.URL.Path != "/api/auth/saml/discover" || r.Header.Get("Authorization") != "" || r.Header.Get("Cache-Control") != "no-store" {
			t.Error("discovery must be anonymous, uncached and selected-origin only")
		}
		_, _ = w.Write([]byte(`{"companyCode":"iteggs"}`))
	}))
	defer server.Close()
	company, err := NewClient(server.URL, true, logging.Default()).DiscoverSAMLCompany(context.Background(), " Gisli@ITEGGS.COM ")
	if err != nil || company != "iteggs" || calls.Load() != 1 {
		t.Fatalf("unexpected discovery: %q %v", company, err)
	}
}

func TestSamlDiscoveryNoMatchFailureAndMalformedResponsePreserveManualLogin(t *testing.T) {
	for _, scenario := range []struct {
		body   string
		status int
	}{
		{`{"companyCode":null}`, 200}, {`{}`, 200}, {`{"companyCode":"https://evil.example"}`, 200},
		{`{"companyCode":7}`, 200}, {`{"companyCode":""}`, 200}, {`internal details`, 503},
	} {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(scenario.status)
			_, _ = w.Write([]byte(scenario.body))
		}))
		company, err := NewClient(server.URL, true, logging.Default()).DiscoverSAMLCompany(context.Background(), "person@iteggs.com")
		server.Close()
		if err == nil || company != "" || !strings.Contains(err.Error(), "normal login") || strings.Contains(err.Error(), "internal details") {
			t.Fatalf("failed discovery hid normal login or accepted invalid data: %q %v", company, err)
		}
	}
}

func TestSamlDiscoveryRejectsDisplayNamesWildcardsAndIncompleteEmailDomains(t *testing.T) {
	for _, input := range []string{"gisli", "gisli@", "gisli@localhost", "Name <gisli@iteggs.com>",
		"gisli@*.iteggs.com", "gisli@iteggs.com.", "gisli@127.0.0.1", "gisli@iteggs.com/path", "a@@iteggs.com"} {
		if _, err := samlEmailDomain(input); err == nil {
			t.Errorf("accepted unsafe email domain: %q", input)
		}
	}
	for input, expected := range map[string]string{"gisli@sub.iteggs.com": "sub.iteggs.com", "gisli@bücher.example": "bücher.example"} {
		if actual, err := samlEmailDomain(input); err != nil || actual != expected {
			t.Errorf("valid exact domain lost: %q %v", actual, err)
		}
	}
}

func TestSamlDiscoveryCancellationStopsTheOperationBeforeReturningACompany(t *testing.T) {
	entered, canceled := make(chan struct{}), make(chan struct{})
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()
		close(canceled)
	}))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() {
		_, err := NewClient(server.URL, true, logging.Default()).DiscoverSAMLCompany(ctx, "gisli@iteggs.com")
		result <- err
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("discovery did not start")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("lost cancellation: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("discovery did not stop")
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("request retained after cancellation")
	}
}
