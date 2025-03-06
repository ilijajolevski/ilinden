// JWT token extraction from requests
//
// Extracts JWT tokens from various sources:
// - URL query parameters
// - Authorization headers
// - Cookies
// - Format normalization

package jwt

import (
	"net/http"
	"sync"

	"github.com/ilijajolevski/ilinden/internal/config"
	"github.com/ilijajolevski/ilinden/pkg/jwtheader"
)

// Extractor handles JWT token extraction from HTTP requests
type Extractor struct {
	opts   jwtheader.ExtractOptions
	config *config.JWTConfig
	mu     sync.RWMutex
}

// NewExtractor creates a new JWT extractor with the provided configuration
func NewExtractor(config *config.JWTConfig) *Extractor {
	return &Extractor{
		opts: jwtheader.ExtractOptions{
			HeaderName: config.HeaderName,
			ParamName:  config.ParamName,
		},
		config: config,
	}
}

// Extract extracts a JWT token from the request
func (e *Extractor) Extract(r *http.Request) (string, error) {
	e.mu.RLock()
	opts := e.opts
	e.mu.RUnlock()

	token, err := jwtheader.FromRequest(r, opts)
	if err != nil {
		if err == jwtheader.ErrNoToken {
			return "", NewTokenRequiredError()
		}
		return "", NewExtractionError(err)
	}

	if !jwtheader.IsValidJWT(token) {
		return "", NewTokenInvalidError()
	}

	return token, nil
}

// UpdateConfig updates the extractor configuration
func (e *Extractor) UpdateConfig(config *config.JWTConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.opts.HeaderName = config.HeaderName
	e.opts.ParamName = config.ParamName
	e.config = config
}

// GetConfig returns the current JWT configuration
func (e *Extractor) GetConfig() *config.JWTConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	return e.config
}

// FromRequest is a convenience function that creates a new extractor and extracts a token
func FromRequest(r *http.Request, config *config.JWTConfig) (string, error) {
	extractor := NewExtractor(config)
	return extractor.Extract(r)
}