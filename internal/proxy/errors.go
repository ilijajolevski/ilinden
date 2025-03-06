// Proxy error handling
package proxy

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// ProxyError represents a proxy-specific error
type ProxyError struct {
	Code       int
	Message    string
	Err        error
	RetryAfter time.Duration
	LogFields  map[string]interface{}
}

// NewProxyError creates a new proxy error
func NewProxyError(code int, message string, err error) *ProxyError {
	return &ProxyError{
		Code:      code,
		Message:   message,
		Err:       err,
		LogFields: make(map[string]interface{}),
	}
}

// Error implements the error interface
func (e *ProxyError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *ProxyError) Unwrap() error {
	return e.Err
}

// WithRetry adds retry information to the error
func (e *ProxyError) WithRetry(retryAfter time.Duration) *ProxyError {
	e.RetryAfter = retryAfter
	return e
}

// WithField adds a log field to the error
func (e *ProxyError) WithField(key string, value interface{}) *ProxyError {
	e.LogFields[key] = value
	return e
}

// WriteResponse writes the error response to the HTTP writer
func (e *ProxyError) WriteResponse(w http.ResponseWriter) {
	// Set status code
	w.WriteHeader(e.Code)
	
	// Set retry header if needed
	if e.RetryAfter > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(int(e.RetryAfter.Seconds())))
	}
	
	// Write error message
	w.Write([]byte(e.Message))
}

// Common error types
var (
	ErrOriginTimeout  = NewProxyError(http.StatusGatewayTimeout, "Origin server timeout", errors.New("origin timeout"))
	ErrOriginRefused  = NewProxyError(http.StatusBadGateway, "Origin server connection refused", errors.New("connection refused"))
	ErrRateLimited    = NewProxyError(http.StatusTooManyRequests, "Rate limit exceeded", errors.New("rate limit"))
	ErrCircuitOpen    = NewProxyError(http.StatusServiceUnavailable, "Service temporarily unavailable", errors.New("circuit open"))
	ErrMalformedURL   = NewProxyError(http.StatusBadRequest, "Malformed URL", errors.New("malformed URL"))
	ErrUnknownService = NewProxyError(http.StatusNotFound, "Unknown service", errors.New("unknown service"))
)