// JWT extraction from headers/params
//
// Generic JWT extraction utilities:
// - Common header patterns
// - Query parameter extraction
// - Format detection
// - Bearer token handling

package jwtheader

import (
	"errors"
	"net/http"
	"strings"
)

const (
	// Common token prefixes
	BearerPrefix = "Bearer "
	JWTPrefix    = "JWT "
)

var (
	// Error definitions
	ErrNoToken          = errors.New("no token found")
	ErrInvalidTokenType = errors.New("invalid token type")
)

// Options for token extraction
type ExtractOptions struct {
	HeaderName string
	ParamName  string
}

// DefaultOptions creates default extraction options
func DefaultOptions() ExtractOptions {
	return ExtractOptions{
		HeaderName: "Authorization",
		ParamName:  "token",
	}
}

// FromHeader extracts a JWT token from the request header
func FromHeader(r *http.Request, headerName string) (string, error) {
	auth := r.Header.Get(headerName)
	if auth == "" {
		return "", ErrNoToken
	}

	// Check for a Bearer or JWT token
	if strings.HasPrefix(auth, BearerPrefix) {
		return strings.TrimPrefix(auth, BearerPrefix), nil
	}
	
	if strings.HasPrefix(auth, JWTPrefix) {
		return strings.TrimPrefix(auth, JWTPrefix), nil
	}
	
	// If no prefix is found but the header exists, just return it as is
	return auth, nil
}

// FromQuery extracts a JWT token from the request query parameters
func FromQuery(r *http.Request, paramName string) (string, error) {
	token := r.URL.Query().Get(paramName)
	if token == "" {
		return "", ErrNoToken
	}
	
	return token, nil
}

// FromRequest extracts a JWT token from a request using the provided options
// It tries the header first, then falls back to query parameters
func FromRequest(r *http.Request, opts ExtractOptions) (string, error) {
	// Try header first
	token, err := FromHeader(r, opts.HeaderName)
	if err == nil {
		return token, nil
	}
	
	if err != ErrNoToken {
		return "", err
	}
	
	// Try query parameter
	return FromQuery(r, opts.ParamName)
}

// IsValidJWT performs basic validation on a JWT token string
func IsValidJWT(token string) bool {
	// A JWT token consists of three parts separated by dots
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	
	// Each part should be non-empty
	for _, part := range parts {
		if part == "" {
			return false
		}
	}
	
	return true
}