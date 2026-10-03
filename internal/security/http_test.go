package security

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type countingTransport struct{ calls int }

func (c *countingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	c.calls++
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("ok"))}, nil
}
func TestTransportRejectsUnsafeOriginsBeforeSending(t *testing.T) {
	for _, base := range []string{"http://remote.example", "https://user:pass@example.com", "https://example.com/api", "https://example.com?x=1"} {
		inner := &countingTransport{}
		request, _ := http.NewRequest("POST", "https://example.com/api/login", strings.NewReader("password"))
		_, err := (Transport{Base: inner, Origin: base}).RoundTrip(request)
		if err == nil || inner.calls != 0 {
			t.Fatalf("unsafe %s reached network", base)
		}
	}
	inner := &countingTransport{}
	request, _ := http.NewRequest("GET", "https://example.com/api", nil)
	if _, err := (Transport{Base: inner, Origin: "https://EXAMPLE.com:443/"}).RoundTrip(request); err != nil || inner.calls != 1 {
		t.Fatalf("valid origin failed: %v", err)
	}
}
func TestRedirectRejectsDowngradeAndDifferentHost(t *testing.T) {
	first, _ := http.NewRequest("GET", "https://example.com/file", nil)
	for _, target := range []string{"http://example.com/file", "https://other.example/file", "https://user@example.com/file"} {
		next, _ := http.NewRequest("GET", target, nil)
		if Redirect(next, []*http.Request{first}) == nil {
			t.Fatal(target)
		}
	}
	next := &http.Request{URL: &url.URL{Scheme: "https", Host: "example.com", Path: "/next"}}
	if err := Redirect(next, []*http.Request{first}); err != nil {
		t.Fatal(err)
	}
}
func TestReadLimitNeverReturnsPartialOversizedJSON(t *testing.T) {
	data, err := ReadLimit(strings.NewReader(`{"ok":true}`+strings.Repeat(" ", 100)), 10)
	if err == nil || data != nil {
		t.Fatal("oversized data returned")
	}
	data, err = ReadLimit(strings.NewReader("ok"), 2)
	if err != nil || string(data) != "ok" {
		t.Fatal(err)
	}
}
