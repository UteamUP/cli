package client

import (
	"context"
	"math"
	"math/rand/v2"
	"net"
	"strings"
	"time"

	clierrors "github.com/uteamup/cli/internal/errors"
	"github.com/uteamup/cli/internal/logging"
)

// RetryOptions configures the retry behavior.
type RetryOptions struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultRetryOptions returns the standard retry configuration.
func DefaultRetryOptions() RetryOptions {
	return RetryOptions{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   10 * time.Second,
	}
}

// RetryWithBackoff executes fn with exponential backoff on retryable errors.
func RetryWithBackoff(ctx context.Context, logger *logging.Logger, operation string, opts RetryOptions, fn func() error) error {
	var lastErr error

	for attempt := 1; attempt <= opts.MaxRetries+1; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if attempt > opts.MaxRetries {
			logger.Error("%s failed after %d retries: %v", operation, opts.MaxRetries, lastErr)
			return lastErr
		}

		if !IsRetryable(lastErr) {
			logger.Error("%s failed with non-retryable error: %v", operation, lastErr)
			return lastErr
		}

		delay := CalculateBackoff(attempt, opts.BaseDelay, opts.MaxDelay)
		logger.Warn("%s failed (attempt %d/%d), retrying in %v: %v",
			operation, attempt, opts.MaxRetries+1, delay, lastErr)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return lastErr
}

// CalculateBackoff computes delay with exponential backoff and +/-20% jitter.
func CalculateBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	delay := float64(baseDelay) * math.Pow(2, float64(attempt-1))
	if delay > float64(maxDelay) {
		delay = float64(maxDelay)
	}
	// Add +/-20% jitter
	jitter := delay * 0.2 * (rand.Float64()*2 - 1)
	return time.Duration(delay + jitter)
}

// IsRetryable determines if an error should be retried.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Network errors
	var netErr net.Error
	if ok := isNetError(err, &netErr); ok {
		return true
	}

	// Check for specific network error strings
	msg := err.Error()
	networkErrors := []string{"connection refused", "connection reset", "no such host", "i/o timeout", "timeout"}
	for _, ne := range networkErrors {
		if strings.Contains(strings.ToLower(msg), ne) {
			return true
		}
	}

	// HTTP 429 or 5xx
	var apiErr *clierrors.APIError
	if isAPIError(err, &apiErr) {
		return apiErr.StatusCode == 429 || apiErr.StatusCode >= 500
	}

	return false
}

func isNetError(err error, target *net.Error) bool {
	for err != nil {
		if ne, ok := err.(net.Error); ok {
			*target = ne
			return true
		}
		if u, ok := err.(interface{ Unwrap() error }); ok {
			err = u.Unwrap()
		} else {
			return false
		}
	}
	return false
}

func isAPIError(err error, target **clierrors.APIError) bool {
	for err != nil {
		if ae, ok := err.(*clierrors.APIError); ok {
			*target = ae
			return true
		}
		if u, ok := err.(interface{ Unwrap() error }); ok {
			err = u.Unwrap()
		} else {
			return false
		}
	}
	return false
}
