// Cache interface definition
//
// Generic cache interface:
// - Get/Set/Delete operations
// - TTL support
// - Stats collection
// - Key value store abstraction

package cache

import (
	"time"
)

// Cache defines a generic caching interface
type Cache interface {
	// Get retrieves a value from the cache
	Get(key Key) (interface{}, bool)
	
	// Set stores a value in the cache with an optional TTL
	Set(key Key, value interface{}, ttl time.Duration)
	
	// Delete removes a value from the cache
	Delete(key Key)
	
	// Clear removes all values from the cache
	Clear()
	
	// Size returns the number of items in the cache
	Size() int
	
	// Stats returns cache statistics
	Stats() Stats
}

// Stats represents cache performance statistics
type Stats struct {
	Hits        uint64
	Misses      uint64
	Size        int
	Evictions   uint64
	Expirations uint64
}

// Factory defines a function that creates a new cache
type Factory func() Cache

// NewMemory creates a new memory cache with default options
func NewMemory() Cache {
	return NewMemoryWithOptions(MemoryOptions{
		MaxSize:   10000,
		ShardSize: 16,
	})
}

// Options configures a cache instance
type Options struct {
	Type        string        // "memory" or "redis"
	TTL         time.Duration // Default TTL for cache entries
	MaxSize     int           // Maximum number of cache entries (for memory cache)
	ShardSize   int           // Number of shards for memory cache
	UseRedis    bool          // Whether to use Redis
	RedisConfig interface{}   // Redis configuration
}

// NewCache creates a new cache with the given options
func NewCache(options Options) Cache {
	if options.UseRedis && options.RedisConfig != nil {
		// Redis cache would be implemented here
		// For now, use memory cache as fallback
		return NewMemory()
	}
	
	return NewMemoryWithOptions(MemoryOptions{
		MaxSize:   options.MaxSize,
		ShardSize: options.ShardSize,
	})
}