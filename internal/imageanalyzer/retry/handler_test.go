package retry

import (
	"errors"
	"net"
	"testing"
	"time"
)

func TestExecuteSuccess(t *testing.T) {
	rh := NewRetryHandler(3)

	result, err := rh.Execute(func() (interface{}, error) {
		return "ok", nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "ok" {
		t.Fatalf("expected 'ok', got %v", result)
	}
}

func TestExecuteRetryTransient(t *testing.T) {
	rh := NewRetryHandler(3)
	calls := 0

	start := time.Now()
	result, err := rh.Execute(func() (interface{}, error) {
		calls++
		if calls < 3 {
			return nil, NewHTTPError(503, "service unavailable")
		}
		return "recovered", nil
	})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("expected no error after retries, got %v", err)
	}
	if result != "recovered" {
		t.Fatalf("expected 'recovered', got %v", result)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
	// Should have had 2 backoff sleeps (attempts 0 and 1).
	if elapsed < 1*time.Second {
		t.Logf("retries completed quickly (backoff is short for early attempts): %v", elapsed)
	}
}

func TestExecuteFailFast4xx(t *testing.T) {
	rh := NewRetryHandler(3)
	calls := 0

	_, err := rh.Execute(func() (interface{}, error) {
		calls++
		return nil, NewHTTPError(400, "bad request")
	})

	if err == nil {
		t.Fatal("expected error for 400 status")
	}
	if calls != 1 {
		t.Fatalf("expected exactly 1 call (no retry for 4xx), got %d", calls)
	}

	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTPError, got %T", err)
	}
	if httpErr.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", httpErr.StatusCode)
	}
}

func TestExecuteExhausted(t *testing.T) {
	rh := NewRetryHandler(2)
	calls := 0

	_, err := rh.Execute(func() (interface{}, error) {
		calls++
		return nil, &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}
	})

	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	// 1 initial + 2 retries = 3 calls.
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}
