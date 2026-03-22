package errors

import "fmt"

// AuthError represents authentication failures.
type AuthError struct {
	Message string
	Cause   error
}

func (e *AuthError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("auth error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("auth error: %s", e.Message)
}

func (e *AuthError) Unwrap() error { return e.Cause }

// NewAuthError creates an AuthError.
func NewAuthError(msg string, cause error) *AuthError {
	return &AuthError{Message: msg, Cause: cause}
}

// APIError represents backend API failures.
type APIError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d %s: %s", e.StatusCode, e.Status, e.Body)
}

// NewAPIError creates an APIError.
func NewAPIError(statusCode int, status, body string) *APIError {
	return &APIError{StatusCode: statusCode, Status: status, Body: body}
}

// ConfigError represents configuration problems.
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("config error [%s]: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("config error: %s", e.Message)
}

// NewConfigError creates a ConfigError.
func NewConfigError(field, msg string) *ConfigError {
	return &ConfigError{Field: field, Message: msg}
}

// ValidationError represents input validation failures.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error [%s]: %s", e.Field, e.Message)
}

// NewValidationError creates a ValidationError.
func NewValidationError(field, msg string) *ValidationError {
	return &ValidationError{Field: field, Message: msg}
}

// NotAuthenticatedError is returned when a command requires auth but no token exists.
type NotAuthenticatedError struct{}

func (e *NotAuthenticatedError) Error() string {
	return "not authenticated — run \"uteamup login\" or \"ut login\" first"
}
