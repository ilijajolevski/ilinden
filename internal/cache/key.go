// Cache key generation
//
// Strategies for generating cache keys:
// - URL-based keys
// - Header influence
// - Query parameter normalization
// - Versioning support

package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"strings"
)

// Key represents a cache key
type Key string

// FromString creates a cache key from a string
func FromString(s string) Key {
	return Key(s)
}

// FromURL creates a cache key from a URL
func FromURL(url string) Key {
	return Key("url:" + url)
}

// FromRequest creates a cache key from an HTTP request
func FromRequest(r *http.Request, opts ...KeyOption) Key {
	options := defaultKeyOptions()
	for _, opt := range opts {
		opt(&options)
	}

	// Start with the path
	key := r.URL.Path

	// Add query parameters if not ignored
	if len(r.URL.RawQuery) > 0 && !options.ignoreQuery {
		if options.normalizeQuery {
			key += "?" + normalizeQuery(r.URL.Query())
		} else {
			key += "?" + r.URL.RawQuery
		}
	}

	// Add method if not ignored
	if !options.ignoreMethod {
		key = r.Method + ":" + key
	}

	// Add considered headers
	if len(options.headers) > 0 {
		headerStr := ""
		for _, header := range options.headers {
			if val := r.Header.Get(header); val != "" {
				headerStr += header + "=" + val + ";"
			}
		}
		if headerStr != "" {
			key += "|" + headerStr
		}
	}

	// Add prefix
	key = options.prefix + key

	// Hash if needed
	if options.hash {
		return hashKey(key)
	}

	return Key(key)
}

// KeyOption configures key generation
type KeyOption func(*keyOptions)

// keyOptions represents options for key generation
type keyOptions struct {
	prefix        string
	ignoreQuery   bool
	ignoreMethod  bool
	normalizeQuery bool
	hash          bool
	headers       []string
}

// defaultKeyOptions returns default key options
func defaultKeyOptions() keyOptions {
	return keyOptions{
		prefix:        "cache:",
		ignoreQuery:   false,
		ignoreMethod:  false,
		normalizeQuery: true,
		hash:          false,
		headers:       []string{},
	}
}

// WithPrefix adds a prefix to the key
func WithPrefix(prefix string) KeyOption {
	return func(o *keyOptions) {
		o.prefix = prefix
	}
}

// IgnoreQuery ignores query parameters in the key
func IgnoreQuery() KeyOption {
	return func(o *keyOptions) {
		o.ignoreQuery = true
	}
}

// IgnoreMethod ignores the HTTP method in the key
func IgnoreMethod() KeyOption {
	return func(o *keyOptions) {
		o.ignoreMethod = true
	}
}

// WithHash hashes the key
func WithHash() KeyOption {
	return func(o *keyOptions) {
		o.hash = true
	}
}

// WithHeaders adds headers to consider in the key
func WithHeaders(headers ...string) KeyOption {
	return func(o *keyOptions) {
		o.headers = append(o.headers, headers...)
	}
}

// DisableQueryNormalization disables query parameter normalization
func DisableQueryNormalization() KeyOption {
	return func(o *keyOptions) {
		o.normalizeQuery = false
	}
}

// normalizeQuery normalizes query parameters for consistent keys
func normalizeQuery(q map[string][]string) string {
	if len(q) == 0 {
		return ""
	}

	// Sort query keys for consistent ordering
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build normalized query string
	var parts []string
	for _, k := range keys {
		values := q[k]
		sort.Strings(values) // Sort values for consistent ordering
		for _, v := range values {
			parts = append(parts, k+"="+v)
		}
	}

	return strings.Join(parts, "&")
}

// hashKey hashes a key string
func hashKey(key string) Key {
	hash := sha256.Sum256([]byte(key))
	return Key(hex.EncodeToString(hash[:]))
}