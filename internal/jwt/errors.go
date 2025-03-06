// JWT error definitions
package jwt

import (
	"errors"
	"fmt"
	"net/http"
)

// JWT-specific errors
var (
	ErrTokenRequired      = errors.New("JWT token is required")
	ErrTokenInvalid       = errors.New("JWT token is invalid")
	ErrTokenExpired       = errors.New("JWT token has expired")
	ErrTokenUnsupported   = errors.New("JWT token uses an unsupported algorithm")
	ErrPlayerIDMissing    = errors.New("player ID is missing in the token")
	ErrExtraction         = errors.New("failed to extract JWT token")
	ErrValidation         = errors.New("JWT token validation failed")
)

// TokenError represents a JWT token error with an HTTP status code
type TokenError struct {
	Err        error
	StatusCode int
	Message    string
}

// NewTokenError creates a new token error
func NewTokenError(err error, statusCode int, message string) *TokenError {
	return &TokenError{
		Err:        err,
		StatusCode: statusCode,
		Message:    message,
	}
}

// Error returns the error message
func (e *TokenError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

// Unwrap returns the underlying error
func (e *TokenError) Unwrap() error {
	return e.Err
}

// Common token errors
func NewTokenRequiredError() *TokenError {
	return NewTokenError(
		ErrTokenRequired,
		http.StatusUnauthorized,
		"authentication token is required",
	)
}

func NewTokenInvalidError() *TokenError {
	return NewTokenError(
		ErrTokenInvalid,
		http.StatusUnauthorized,
		"authentication token is invalid",
	)
}

func NewTokenExpiredError() *TokenError {
	return NewTokenError(
		ErrTokenExpired,
		http.StatusUnauthorized,
		"authentication token has expired",
	)
}

func NewExtractionError(err error) *TokenError {
	return NewTokenError(
		fmt.Errorf("%w: %v", ErrExtraction, err),
		http.StatusBadRequest,
		"failed to extract authentication token",
	)
}

func NewValidationError(err error) *TokenError {
	return NewTokenError(
		fmt.Errorf("%w: %v", ErrValidation, err),
		http.StatusUnauthorized,
		"authentication token validation failed",
	)
}