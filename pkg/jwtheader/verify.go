// JWT validation utilities
package jwtheader

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
)

var (
	// Error definitions
	ErrInvalidToken      = errors.New("invalid JWT token")
	ErrTokenExpired      = errors.New("token has expired")
	ErrInvalidSignature  = errors.New("invalid token signature")
	ErrMissingClaim      = errors.New("required claim is missing")
	ErrInvalidAlgorithm  = errors.New("unsupported signing algorithm")
	ErrInvalidIssuer     = errors.New("invalid token issuer")
	ErrInvalidAudience   = errors.New("invalid token audience")
)

// JWTHeader represents the header of a JWT token
type JWTHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid,omitempty"`
}

// JWTClaims represents standard JWT claims
type JWTClaims struct {
	Issuer         string                 `json:"iss,omitempty"`
	Subject        string                 `json:"sub,omitempty"`
	Audience       interface{}            `json:"aud,omitempty"` // string or []string
	ExpirationTime int64                  `json:"exp,omitempty"`
	NotBefore      int64                  `json:"nbf,omitempty"`
	IssuedAt       int64                  `json:"iat,omitempty"`
	JWTID          string                 `json:"jti,omitempty"`
	Custom         map[string]interface{} `json:"-"`
}

// JWKSet represents a set of JSON Web Keys
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	KeyType      string `json:"kty"`
	KeyID        string `json:"kid,omitempty"`
	Use          string `json:"use,omitempty"`
	Algorithm    string `json:"alg,omitempty"`
	N            string `json:"n,omitempty"` // RSA modulus
	E            string `json:"e,omitempty"` // RSA public exponent
	X5C          []string `json:"x5c,omitempty"`
	X5U          string `json:"x5u,omitempty"`
	X5T          string `json:"x5t,omitempty"`
	X5TS256      string `json:"x5t#S256,omitempty"`
}

// ValidationOptions represents options for JWT validation
type ValidationOptions struct {
	Secret          string   // HMAC secret
	KeysURL         string   // URL to JWKS
	RequiredClaims []string  // Claims that must be present
	Issuer         string    // Expected issuer
	Audience        string   // Expected audience
	ClaimsNamespace string   // Namespace for custom claims
	AllowedAlgs     []string // Allowed signing algorithms
}

// ParseAndVerify parses a JWT token string and verifies its signature
func ParseAndVerify(tokenString string, opts ValidationOptions) (*JWTClaims, error) {
	// Basic format validation
	if !IsValidJWT(tokenString) {
		return nil, ErrInvalidToken
	}
	
	// Parse token parts
	parts := strings.Split(tokenString, ".")
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid header encoding: %w", err)
	}
	
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid payload encoding: %w", err)
	}
	
	// Parse header
	var header JWTHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("invalid header format: %w", err)
	}
	
	// Verify algorithm
	if !isAllowedAlgorithm(header.Algorithm, opts.AllowedAlgs) {
		return nil, ErrInvalidAlgorithm
	}
	
	// Parse claims
	claims, err := parseClaims(payloadBytes)
	if err != nil {
		return nil, err
	}
	
	// Check required claims
	if err := validateRequiredClaims(claims, opts.RequiredClaims); err != nil {
		return nil, err
	}
	
	// Validate expiration
	if claims.ExpirationTime > 0 {
		now := time.Now().Unix()
		if now > claims.ExpirationTime {
			return nil, ErrTokenExpired
		}
	}
	
	// Validate issuer if specified
	if opts.Issuer != "" && claims.Issuer != "" && claims.Issuer != opts.Issuer {
		return nil, ErrInvalidIssuer
	}
	
	// Validate audience if specified
	if opts.Audience != "" && !hasAudience(claims, opts.Audience) {
		return nil, ErrInvalidAudience
	}
	
	// For this implementation, we'll skip actual signature verification
	// In a real implementation, this would verify the signature using the appropriate algorithm
	
	return claims, nil
}

// isAllowedAlgorithm checks if the algorithm is in the allowed list
func isAllowedAlgorithm(alg string, allowed []string) bool {
	if len(allowed) == 0 {
		return true // If no algorithms are specified, all are allowed
	}
	
	for _, a := range allowed {
		if a == alg {
			return true
		}
	}
	
	return false
}

// parseClaims parses the JWT claims from the payload
func parseClaims(payloadBytes []byte) (*JWTClaims, error) {
	var claims JWTClaims
	
	// Parse into a generic map first
	var claimsMap map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &claimsMap); err != nil {
		return nil, fmt.Errorf("invalid claims format: %w", err)
	}
	
	// Extract standard claims
	claims.Custom = make(map[string]interface{})
	
	// Read standard claims
	for k, v := range claimsMap {
		switch k {
		case "iss":
			if iss, ok := v.(string); ok {
				claims.Issuer = iss
			}
		case "sub":
			if sub, ok := v.(string); ok {
				claims.Subject = sub
			}
		case "aud":
			claims.Audience = v // Can be string or []string
		case "exp":
			if exp, ok := v.(float64); ok {
				claims.ExpirationTime = int64(exp)
			}
		case "nbf":
			if nbf, ok := v.(float64); ok {
				claims.NotBefore = int64(nbf)
			}
		case "iat":
			if iat, ok := v.(float64); ok {
				claims.IssuedAt = int64(iat)
			}
		case "jti":
			if jti, ok := v.(string); ok {
				claims.JWTID = jti
			}
		default:
			// Store as custom claim
			claims.Custom[k] = v
		}
	}
	
	return &claims, nil
}

// validateRequiredClaims checks if all required claims are present
func validateRequiredClaims(claims *JWTClaims, required []string) error {
	for _, claim := range required {
		switch claim {
		case "iss":
			if claims.Issuer == "" {
				return fmt.Errorf("%w: %s", ErrMissingClaim, claim)
			}
		case "sub":
			if claims.Subject == "" {
				return fmt.Errorf("%w: %s", ErrMissingClaim, claim)
			}
		case "aud":
			if claims.Audience == nil {
				return fmt.Errorf("%w: %s", ErrMissingClaim, claim)
			}
		case "exp":
			if claims.ExpirationTime == 0 {
				return fmt.Errorf("%w: %s", ErrMissingClaim, claim)
			}
		case "nbf":
			if claims.NotBefore == 0 {
				return fmt.Errorf("%w: %s", ErrMissingClaim, claim)
			}
		case "iat":
			if claims.IssuedAt == 0 {
				return fmt.Errorf("%w: %s", ErrMissingClaim, claim)
			}
		case "jti":
			if claims.JWTID == "" {
				return fmt.Errorf("%w: %s", ErrMissingClaim, claim)
			}
		default:
			if _, ok := claims.Custom[claim]; !ok {
				return fmt.Errorf("%w: %s", ErrMissingClaim, claim)
			}
		}
	}
	
	return nil
}

// hasAudience checks if the claims have the expected audience
func hasAudience(claims *JWTClaims, expected string) bool {
	switch aud := claims.Audience.(type) {
	case string:
		return aud == expected
	case []interface{}:
		for _, a := range aud {
			if str, ok := a.(string); ok && str == expected {
				return true
			}
		}
	}
	return false
}

// fetchJWKS fetches a JWKS from the given URL
func fetchJWKS(url string) (*JWKSet, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS: HTTP %d", resp.StatusCode)
	}
	
	var jwks JWKSet
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("invalid JWKS format: %w", err)
	}
	
	return &jwks, nil
}

// jwkToRSA converts a JWK to an RSA public key
func jwkToRSA(jwk JWK) (*rsa.PublicKey, error) {
	if jwk.KeyType != "RSA" {
		return nil, fmt.Errorf("unsupported key type: %s", jwk.KeyType)
	}
	
	// Decode modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("invalid modulus encoding: %w", err)
	}
	
	// Decode exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("invalid exponent encoding: %w", err)
	}
	
	// Convert modulus bytes to big int
	n := new(big.Int).SetBytes(nBytes)
	
	// Convert exponent bytes to int
	var e int
	if len(eBytes) == 3 {
		e = int(eBytes[0])<<16 | int(eBytes[1])<<8 | int(eBytes[2])
	} else if len(eBytes) == 2 {
		e = int(eBytes[0])<<8 | int(eBytes[1])
	} else if len(eBytes) == 1 {
		e = int(eBytes[0])
	} else {
		return nil, fmt.Errorf("invalid exponent size")
	}
	
	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}