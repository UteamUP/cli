package security

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

const MaxResponseBytes int64 = 4 * 1024 * 1024

// Origin returns the canonical clean HTTPS origin used to bind a session.
func Origin(raw string) (string, error) {
	u, err := url.ParseRequestURI(raw)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return "", fmt.Errorf("API base URL must be a clean HTTPS origin")
	}
	host := strings.ToLower(u.Hostname())
	port := u.Port()
	if port != "" && port != "443" {
		host = net.JoinHostPort(host, port)
	} else if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	return "https://" + host, nil
}

// Transport enforces origin validation before any credential or body leaves the process.
type Transport struct {
	Base   http.RoundTripper
	Origin string
}

func (t Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	origin, err := Origin(t.Origin)
	if err != nil {
		return nil, err
	}
	target, err := Origin(req.URL.Scheme + "://" + req.URL.Host)
	if err != nil || target != origin || req.URL.User != nil {
		return nil, fmt.Errorf("request origin differs from the selected HTTPS backend")
	}
	return t.Base.RoundTrip(req)
}

// Redirect refuses to move credentials, POST bodies, or downloads across a trust boundary.
func Redirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 5 {
		return fmt.Errorf("redirect limit exceeded")
	}
	origin, err := Origin(req.URL.Scheme + "://" + req.URL.Host)
	if err != nil || req.URL.User != nil {
		return fmt.Errorf("redirect must use HTTPS without credentials")
	}
	first, err := Origin(via[0].URL.Scheme + "://" + via[0].URL.Host)
	if err != nil || origin != first {
		return fmt.Errorf("cross-origin redirect rejected")
	}
	return nil
}

func ReadAll(r io.Reader) ([]byte, error) { return ReadLimit(r, MaxResponseBytes) }
func ReadLimit(r io.Reader, limit int64) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("response exceeds %d byte limit", limit)
	}
	return b, nil
}
