// Package retry provides retry logic with exponential backoff for transient errors.
package retry

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"time"
)

// HTTPError represents an HTTP error with a status code.
type HTTPError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// NewHTTPError creates a new HTTPError with the given status code and message.
func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{StatusCode: statusCode, Message: message}
}

// RetryHandler retries transient errors with exponential backoff and jitter.
type RetryHandler struct {
	maxRetries int
}

// NewRetryHandler creates a RetryHandler with the given maximum number of retries.
func NewRetryHandler(maxRetries int) *RetryHandler {
	if maxRetries < 0 {
		maxRetries = 0
	}
	return &RetryHandler{maxRetries: maxRetries}
}

// Execute calls fn and retries on transient errors with exponential backoff.
// Non-transient errors (4xx except 429) fail immediately.
func (rh *RetryHandler) Execute(fn func() (interface{}, error)) (interface{}, error) {
	var lastErr error

	for attempt := 0; attempt <= rh.maxRetries; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		if !rh.isTransient(err) {
			return nil, err
		}

		if attempt >= rh.maxRetries {
			log.Printf("retry: exhausted after %d attempts, error=%v", attempt+1, err)
			return nil, lastErr
		}

		rh.backoff(attempt, err)
	}

	return nil, lastErr
}

// isTransient checks whether the error is transient and should be retried.
func (rh *RetryHandler) isTransient(err error) bool {
	// Check for net.Error (timeout, temporary).
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			return true
		}
	}

	// Check for connection refused and other OpError.
	if _, ok := err.(*net.OpError); ok {
		return true
	}

	// Check for HTTPError with transient status codes.
	if httpErr, ok := err.(*HTTPError); ok {
		switch httpErr.StatusCode {
		case 429, 500, 502, 503, 504:
			return true
		default:
			// 4xx (not 429) are permanent — fail fast.
			if httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 {
				return false
			}
		}
	}

	return false
}

// backoff sleeps with exponential backoff plus random jitter.
func (rh *RetryHandler) backoff(attempt int, err error) {
	baseDelay := math.Pow(2, float64(attempt)) // 1, 2, 4, 8, ...
	jitter := rand.Float64() * baseDelay * 0.5 // 0 to baseDelay*0.5
	delay := time.Duration((baseDelay + jitter) * float64(time.Second))
	log.Printf("retry: attempt %d/%d, delay=%.2fs, error=%v", attempt+1, rh.maxRetries, delay.Seconds(), err)
	time.Sleep(delay)
}
