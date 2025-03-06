// JWT claims handling
//
// Custom claims extraction and processing:
// - Standard claims support
// - Custom claim validation
// - Player ID extraction
// - Role/permission handling

package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/ilijajolevski/ilinden/pkg/jwtheader"
)

// Claims wraps the standard JWT claims and adds application-specific functionality
type Claims struct {
	*jwtheader.JWTClaims
	namespace string
}

// NewClaims creates a new Claims instance from JWTClaims
func NewClaims(claims *jwtheader.JWTClaims, namespace string) *Claims {
	return &Claims{
		JWTClaims: claims,
		namespace: namespace,
	}
}

// GetPlayerID extracts the player ID from the claims
func (c *Claims) GetPlayerID() (string, error) {
	// Try to get from subject claim first
	if c.Subject != "" {
		return c.Subject, nil
	}
	
	// Try to get from custom playerId claim
	if c.namespace != "" {
		nsKey := c.namespace + "playerId"
		if playerID, ok := c.Custom[nsKey]; ok {
			if id, ok := playerID.(string); ok && id != "" {
				return id, nil
			}
		}
	}
	
	// Try standard custom playerId claim
	if playerID, ok := c.Custom["playerId"]; ok {
		if id, ok := playerID.(string); ok && id != "" {
			return id, nil
		}
	}
	
	return "", errors.New("player ID not found in token")
}

// GetCustomClaim retrieves a custom claim value, considering the namespace
func (c *Claims) GetCustomClaim(name string) (interface{}, bool) {
	// Try namespaced claim first
	if c.namespace != "" {
		nsKey := c.namespace + name
		if val, ok := c.Custom[nsKey]; ok {
			return val, true
		}
	}
	
	// Fall back to standard claim
	val, ok := c.Custom[name]
	return val, ok
}

// GetStringClaim retrieves a string custom claim
func (c *Claims) GetStringClaim(name string) (string, bool) {
	val, ok := c.GetCustomClaim(name)
	if !ok {
		return "", false
	}
	
	str, ok := val.(string)
	return str, ok
}

// HasRole checks if the token has a specific role
func (c *Claims) HasRole(role string) bool {
	// Try to get roles from custom claim
	roles, ok := c.GetCustomClaim("roles")
	if !ok {
		return false
	}
	
	// Check if roles is a string array
	if rolesArr, ok := roles.([]interface{}); ok {
		for _, r := range rolesArr {
			if rStr, ok := r.(string); ok && rStr == role {
				return true
			}
		}
	}
	
	return false
}

// IsExpired checks if the token is expired
func (c *Claims) IsExpired() bool {
	if c.ExpirationTime == 0 {
		return false // No expiration time means token doesn't expire
	}
	
	now := time.Now().Unix()
	return now > c.ExpirationTime
}

// RemainingValidity returns the remaining validity time of the token in seconds
func (c *Claims) RemainingValidity() int64 {
	if c.ExpirationTime == 0 {
		return 0 // No expiration time
	}
	
	now := time.Now().Unix()
	remaining := c.ExpirationTime - now
	
	if remaining < 0 {
		return 0 // Token already expired
	}
	
	return remaining
}

// String returns a string representation of the claims
func (c *Claims) String() string {
	if c == nil || c.JWTClaims == nil {
		return "<nil>"
	}
	
	return fmt.Sprintf("Subject: %s, Issuer: %s, Expires: %d", 
		c.Subject, c.Issuer, c.ExpirationTime)
}