// Package spinner provides an animated braille spinner for terminal output
// during long-running video upload and processing operations.
package spinner

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Spinner displays an animated braille spinner on stderr with configurable status text.
type Spinner struct {
	frames    []rune
	text      string
	mu        sync.Mutex
	stopCh    chan struct{}
	doneCh    chan struct{}
	running   bool
	startedAt time.Time
}

// New creates a new Spinner with braille animation frames.
func New() *Spinner {
	return &Spinner{
		frames: []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'},
	}
}

// Start begins the spinner animation on stderr. Safe to call multiple times (no-op if running).
func (s *Spinner) Start(text string) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.text = text
	s.running = true
	s.startedAt = time.Now()
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	s.mu.Unlock()

	go func() {
		defer close(s.doneCh)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		frameIdx := 0
		for {
			select {
			case <-s.stopCh:
				return
			case <-ticker.C:
				s.mu.Lock()
				elapsed := FormatElapsed(time.Since(s.startedAt))
				frame := s.frames[frameIdx%len(s.frames)]
				currentText := s.text
				s.mu.Unlock()

				fmt.Fprintf(os.Stderr, "\r%c %s (%s)", frame, currentText, elapsed)
				frameIdx++
			}
		}
	}()
}

// UpdateText changes the status text while the spinner is running.
func (s *Spinner) UpdateText(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.text = text
}

// Stop stops the spinner and clears the line.
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopCh)
	<-s.doneCh

	// Clear the line.
	fmt.Fprintf(os.Stderr, "\r%*s\r", 80, "")
}

// StopWithMessage stops the spinner and prints a final message.
func (s *Spinner) StopWithMessage(msg string) {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopCh)
	<-s.doneCh

	// Clear the line and print the message.
	fmt.Fprintf(os.Stderr, "\r%*s\r", 80, "")
	fmt.Fprintf(os.Stderr, "%s\n", msg)
}

// FormatElapsed returns a human-readable elapsed time string like "12s" or "2m15s".
func FormatElapsed(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	if totalSeconds < 60 {
		return fmt.Sprintf("%ds", totalSeconds)
	}
	if totalSeconds < 3600 {
		minutes := totalSeconds / 60
		seconds := totalSeconds % 60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
}
