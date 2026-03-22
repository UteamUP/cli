package logging

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Level represents log severity.
type Level int

const (
	LevelTrace Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
)

// ParseLevel converts a string to a Level.
func ParseLevel(s string) Level {
	switch strings.ToUpper(s) {
	case "TRACE":
		return LevelTrace
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

func (l Level) String() string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "INFO"
	}
}

// Logger provides structured logging with sensitive data redaction.
type Logger struct {
	level Level
}

// New creates a Logger at the given level.
func New(level Level) *Logger {
	return &Logger{level: level}
}

// Default returns an INFO-level logger.
func Default() *Logger {
	return New(LevelInfo)
}

func (l *Logger) log(level Level, msg string, args ...any) {
	if level < l.level {
		return
	}
	ts := time.Now().UTC().Format("15:04:05")
	formatted := fmt.Sprintf(msg, args...)
	formatted = redact(formatted)
	fmt.Fprintf(os.Stderr, "[%s] %s  %s\n", ts, level, formatted)
}

// Trace logs at TRACE level.
func (l *Logger) Trace(msg string, args ...any) { l.log(LevelTrace, msg, args...) }

// Debug logs at DEBUG level.
func (l *Logger) Debug(msg string, args ...any) { l.log(LevelDebug, msg, args...) }

// Info logs at INFO level.
func (l *Logger) Info(msg string, args ...any) { l.log(LevelInfo, msg, args...) }

// Warn logs at WARN level.
func (l *Logger) Warn(msg string, args ...any) { l.log(LevelWarn, msg, args...) }

// Error logs at ERROR level.
func (l *Logger) Error(msg string, args ...any) { l.log(LevelError, msg, args...) }

// SetLevel changes the log level.
func (l *Logger) SetLevel(level Level) { l.level = level }

// redact removes sensitive data from log messages.
func redact(s string) string {
	sensitivePatterns := []string{"Bearer ", "token", "secret", "password", "apiKey", "api_key"}
	result := s
	for _, pattern := range sensitivePatterns {
		idx := strings.Index(strings.ToLower(result), strings.ToLower(pattern))
		if idx == -1 {
			continue
		}
		// Find the value after the pattern (look for = or : or space followed by the value)
		after := result[idx+len(pattern):]
		for _, sep := range []string{"=", ":", " "} {
			if strings.HasPrefix(after, sep) {
				valueStart := idx + len(pattern) + len(sep)
				valueEnd := findValueEnd(result, valueStart)
				if valueEnd > valueStart {
					result = result[:valueStart] + "[REDACTED]" + result[valueEnd:]
				}
				break
			}
		}
	}
	return result
}

// findValueEnd finds where a value ends (next space, quote boundary, or EOL).
func findValueEnd(s string, start int) int {
	if start >= len(s) {
		return start
	}
	// If value starts with a quote, find closing quote
	if s[start] == '"' || s[start] == '\'' {
		quote := s[start]
		for i := start + 1; i < len(s); i++ {
			if s[i] == quote {
				return i + 1
			}
		}
		return len(s)
	}
	// Otherwise find next whitespace or comma
	for i := start; i < len(s); i++ {
		if s[i] == ' ' || s[i] == ',' || s[i] == '\n' || s[i] == '\t' {
			return i
		}
	}
	return len(s)
}
