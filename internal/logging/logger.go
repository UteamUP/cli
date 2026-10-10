package logging

import (
	"fmt"
	"github.com/uteamup/cli/internal/security"
	"os"
	"regexp"
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
	formatted = SafeDiagnostic(formatted)
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

var capabilityURL = regexp.MustCompile(`(?i)https?://[^\s"'<>]+`)
var credentialValue = regexp.MustCompile(`(?i)(\bBearer\s+|\b(?:token|secret|password|apiKey|api_key)\s*(?:=|:)\s*|\b(?:token|secret|password|apiKey|api_key)\s+)(?:"[^"\r\n]*"|'[^'\r\n]*'|[^\s,&#]+)`)

// redact removes every credential occurrence, including URLs in transport errors.
func redact(s string) string {
	result := capabilityURL.ReplaceAllStringFunc(s, func(raw string) string {
		if index := strings.IndexAny(raw, "?#"); index >= 0 {
			return raw[:index] + "?[REDACTED]"
		}
		return raw
	})
	return credentialValue.ReplaceAllString(result, "${1}[REDACTED]")
}

// SafeDiagnostic protects final command errors as well as logger output.
func SafeDiagnostic(message string) string { return security.SafeText(redact(message)) }
