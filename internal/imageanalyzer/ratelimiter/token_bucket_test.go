package ratelimiter

import (
	"testing"
	"time"
)

func TestAcquireImmediate(t *testing.T) {
	rl := NewTokenBucket(60) // 60 RPM = 1 token/sec, starts with 60 tokens

	start := time.Now()
	rl.Acquire()
	elapsed := time.Since(start)

	if elapsed > 100*time.Millisecond {
		t.Errorf("Acquire with available tokens took too long: %v", elapsed)
	}
}

func TestAcquireRefill(t *testing.T) {
	rl := NewTokenBucket(60)

	// Exhaust all tokens.
	for i := 0; i < 60; i++ {
		rl.mu.Lock()
		rl.tokens -= 1
		rl.mu.Unlock()
	}

	// Wait long enough for at least one token to refill (1 token/sec at 60 RPM).
	time.Sleep(1100 * time.Millisecond)

	start := time.Now()
	rl.Acquire()
	elapsed := time.Since(start)

	if elapsed > 200*time.Millisecond {
		t.Errorf("Acquire after refill should be immediate, took: %v", elapsed)
	}
}

func TestAcquireBlocking(t *testing.T) {
	rl := NewTokenBucket(60) // refill rate = 1 token/sec

	// Exhaust all tokens.
	rl.mu.Lock()
	rl.tokens = 0
	rl.lastRefill = time.Now()
	rl.mu.Unlock()

	start := time.Now()
	rl.Acquire()
	elapsed := time.Since(start)

	// Should have blocked for approximately 1 second waiting for a token.
	if elapsed < 800*time.Millisecond {
		t.Errorf("Acquire should have blocked, but returned in %v", elapsed)
	}
	if elapsed > 3*time.Second {
		t.Errorf("Acquire blocked too long: %v", elapsed)
	}
}
