package spinner

import (
	"testing"
	"time"
)

func TestSpinner_StartStop(t *testing.T) {
	s := New()
	s.Start("Processing...")
	time.Sleep(250 * time.Millisecond)
	s.Stop()
	// No panic, no hang = pass.
}

func TestSpinner_UpdateText(t *testing.T) {
	s := New()
	s.Start("Starting...")
	time.Sleep(150 * time.Millisecond)
	s.UpdateText("Updated text")
	time.Sleep(150 * time.Millisecond)
	s.Stop()
	// No panic = pass.
}

func TestSpinner_StopWithMessage(t *testing.T) {
	s := New()
	s.Start("Working...")
	time.Sleep(150 * time.Millisecond)
	s.StopWithMessage("Done!")
	// No panic = pass.
}

func TestSpinner_DoubleStart(t *testing.T) {
	s := New()
	s.Start("First start")
	s.Start("Second start") // Should be no-op.
	time.Sleep(150 * time.Millisecond)
	s.Stop()
	// No panic = pass.
}

func TestSpinner_StopWithoutStart(t *testing.T) {
	s := New()
	s.Stop() // Should be no-op.
	// No panic = pass.
}

func TestFormatElapsed(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "zero seconds",
			duration: 0 * time.Second,
			expected: "0s",
		},
		{
			name:     "five seconds",
			duration: 5 * time.Second,
			expected: "5s",
		},
		{
			name:     "sixty-five seconds",
			duration: 65 * time.Second,
			expected: "1m5s",
		},
		{
			name:     "one hour one minute one second",
			duration: 3661 * time.Second,
			expected: "1h1m1s",
		},
		{
			name:     "exactly one minute",
			duration: 60 * time.Second,
			expected: "1m0s",
		},
		{
			name:     "exactly one hour",
			duration: 3600 * time.Second,
			expected: "1h0m0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatElapsed(tt.duration)
			if result != tt.expected {
				t.Errorf("FormatElapsed(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}
