package client

import (
	"fmt"
	"testing"
	"time"

	clierrors "github.com/uteamup/cli/internal/errors"
)

func TestCalculateBackoff(t *testing.T) {
	base := 1 * time.Second
	max := 10 * time.Second

	tests := []struct {
		attempt int
		minMS   int64
		maxMS   int64
	}{
		{1, 800, 1200},   // 1s +/- 20%
		{2, 1600, 2400},  // 2s +/- 20%
		{3, 3200, 4800},  // 4s +/- 20%
		{4, 6400, 9600},  // 8s +/- 20%
		{5, 8000, 12000}, // capped at 10s +/- 20%
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tc.attempt), func(t *testing.T) {
			delay := CalculateBackoff(tc.attempt, base, max)
			ms := delay.Milliseconds()
			if ms < tc.minMS || ms > tc.maxMS {
				t.Errorf("attempt %d: delay %dms outside expected range [%d, %d]",
					tc.attempt, ms, tc.minMS, tc.maxMS)
			}
		})
	}
}

func TestCalculateBackoffCapsAtMax(t *testing.T) {
	base := 1 * time.Second
	max := 5 * time.Second

	// Attempt 10 should be capped at max (5s) +/- 20%
	for i := 0; i < 100; i++ {
		delay := CalculateBackoff(10, base, max)
		if delay > 6*time.Second { // 5s + 20%
			t.Errorf("delay %v exceeds max + jitter", delay)
		}
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{"nil error", nil, false},
		{"generic error", fmt.Errorf("something broke"), false},
		{"connection refused", fmt.Errorf("connection refused"), true},
		{"connection reset", fmt.Errorf("connection reset by peer"), true},
		{"timeout", fmt.Errorf("request timeout exceeded"), true},
		{"no such host", fmt.Errorf("no such host"), true},
		{"HTTP 429", clierrors.NewAPIError(429, "Too Many Requests", "rate limited"), true},
		{"HTTP 500", clierrors.NewAPIError(500, "Internal Server Error", ""), true},
		{"HTTP 502", clierrors.NewAPIError(502, "Bad Gateway", ""), true},
		{"HTTP 503", clierrors.NewAPIError(503, "Service Unavailable", ""), true},
		{"HTTP 400", clierrors.NewAPIError(400, "Bad Request", "invalid"), false},
		{"HTTP 401", clierrors.NewAPIError(401, "Unauthorized", ""), false},
		{"HTTP 403", clierrors.NewAPIError(403, "Forbidden", ""), false},
		{"HTTP 404", clierrors.NewAPIError(404, "Not Found", ""), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsRetryable(tc.err)
			if result != tc.retryable {
				t.Errorf("IsRetryable(%v) = %v, want %v", tc.err, result, tc.retryable)
			}
		})
	}
}
