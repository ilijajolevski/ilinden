// JWT validation logic
//
// JWT token validation:
// - Signature verification
// - Claims validation
// - Expiration checking
// - Issuer validation
// - Caching of validation results

package jwt

import (
	"sync"
	"time"

	"github.com/ilijajolevski/ilinden/internal/cache"
	"github.com/ilijajolevski/ilinden/internal/config"
	"github.com/ilijajolevski/ilinden/pkg/jwtheader"
)

// Validator handles JWT token validation
type Validator struct {
	config     *config.JWTConfig
	cache      cache.Cache
	cacheTTL   time.Duration
	validCache bool
	mu         sync.RWMutex
}

// NewValidator creates a new JWT validator with the provided configuration
func NewValidator(config *config.JWTConfig, optionalCache cache.Cache) *Validator {
	v := &Validator{
		config:     config,
		cacheTTL:   5 * time.Minute,
		validCache: optionalCache != nil,
	}

	if optionalCache != nil {
		v.cache = optionalCache
	}

	return v
}

// ValidateToken validates a JWT token and returns the parsed claims
func (v *Validator) ValidateToken(token string) (*Claims, error) {
	v.mu.RLock()
	config := v.config
	useCache := v.validCache
	v.mu.RUnlock()

	// Check cache first if available
	if useCache {
		cachedClaims, found := v.getFromCache(token)
		if found {
			// Check if token has expired since being cached
			if cachedClaims.IsExpired() {
				v.removeFromCache(token)
				return nil, NewTokenExpiredError()
			}
			return cachedClaims, nil
		}
	}

	// Prepare validation options
	opts := jwtheader.ValidationOptions{
		Secret:          config.Secret,
		KeysURL:         config.KeysURL,
		RequiredClaims:  config.RequiredClaims,
		Issuer:          config.Issuer,
		Audience:        config.Audience,
		ClaimsNamespace: config.ClaimsNamespace,
		AllowedAlgs:     config.AllowedAlgs,
	}

	// Validate token
	jwtClaims, err := jwtheader.ParseAndVerify(token, opts)
	if err != nil {
		// Map specific error types
		switch err {
		case jwtheader.ErrTokenExpired:
			return nil, NewTokenExpiredError()
		case jwtheader.ErrInvalidToken, jwtheader.ErrInvalidSignature:
			return nil, NewTokenInvalidError()
		default:
			return nil, NewValidationError(err)
		}
	}

	// Create our claims wrapper
	claims := NewClaims(jwtClaims, config.ClaimsNamespace)

	// Cache valid claims if caching is enabled
	if useCache {
		v.addToCache(token, claims)
	}

	return claims, nil
}

// UpdateConfig updates the validator configuration
func (v *Validator) UpdateConfig(config *config.JWTConfig) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.config = config
}

// getFromCache tries to get claims from the cache
func (v *Validator) getFromCache(token string) (*Claims, bool) {
	if v.cache == nil {
		return nil, false
	}

	key := cache.Key("jwt:token:" + token)
	value, found := v.cache.Get(key)
	if !found {
		return nil, false
	}

	claims, ok := value.(*Claims)
	return claims, ok
}

// addToCache adds claims to the cache
func (v *Validator) addToCache(token string, claims *Claims) {
	if v.cache == nil {
		return
	}

	key := cache.Key("jwt:token:" + token)
	ttl := v.cacheTTL

	// If token has an expiration, use that as TTL instead
	// (minus a small buffer to ensure we don't serve nearly-expired tokens)
	if claims.ExpirationTime > 0 {
		remaining := claims.RemainingValidity()
		if remaining > 0 {
			// Use the lower of the two values
			expTTL := time.Duration(remaining-30) * time.Second
			if expTTL < ttl {
				ttl = expTTL
			}
		}
	}

	v.cache.Set(key, claims, ttl)
}

// removeFromCache removes claims from the cache
func (v *Validator) removeFromCache(token string) {
	if v.cache == nil {
		return
	}

	key := cache.Key("jwt:token:" + token)
	v.cache.Delete(key)
}

// Static convenience method
// ValidateTokenWithConfig validates a token with the provided configuration
func ValidateTokenWithConfig(token string, config *config.JWTConfig) (*Claims, error) {
	validator := NewValidator(config, nil)
	return validator.ValidateToken(token)
}