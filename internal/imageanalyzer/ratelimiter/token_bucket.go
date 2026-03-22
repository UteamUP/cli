// Package ratelimiter provides token-bucket rate limiting for API calls.
package ratelimiter

import (
	"log"
	"sync"
	"time"
)

// TokenBucketRateLimiter controls the rate of API requests using a token bucket algorithm.
type TokenBucketRateLimiter struct {
	rpm        int
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucket creates a new TokenBucketRateLimiter for the given requests per minute.
func NewTokenBucket(requestsPerMinute int) *TokenBucketRateLimiter {
	if requestsPerMinute < 1 {
		requestsPerMinute = 1
	}
	rpm := requestsPerMinute
	return &TokenBucketRateLimiter{
		rpm:        rpm,
		tokens:     float64(rpm),
		maxTokens:  float64(rpm),
		refillRate: float64(rpm) / 60.0,
		lastRefill: time.Now(),
	}
}

// Acquire blocks until a token is available, then consumes one.
func (rl *TokenBucketRateLimiter) Acquire() {
	for {
		rl.mu.Lock()
		rl.refill()
		if rl.tokens >= 1.0 {
			rl.tokens -= 1.0
			remaining := rl.tokens
			rl.mu.Unlock()
			log.Printf("rate_limiter: token acquired, tokens_remaining=%.2f, rpm=%d", remaining, rl.rpm)
			return
		}
		// Calculate wait time until at least one token is available.
		waitDuration := time.Duration((1.0 - rl.tokens) / rl.refillRate * float64(time.Second))
		rl.mu.Unlock()
		log.Printf("rate_limiter: waiting %.3f seconds for token", waitDuration.Seconds())
		time.Sleep(waitDuration)
	}
}

// refill adds tokens based on elapsed time since last refill. Must be called with mu held.
func (rl *TokenBucketRateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	if elapsed <= 0 {
		return
	}
	newTokens := elapsed * rl.refillRate
	rl.tokens += newTokens
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}
	rl.lastRefill = now
}
