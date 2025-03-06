// In-memory cache implementation
//
// LRU-based memory cache:
// - Concurrent access support
// - Size-based eviction
// - TTL-based expiration
// - Memory usage limiting

package cache

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

// MemoryCache implements an in-memory cache
type MemoryCache struct {
	shards    []*memoryShard
	shardMask uint32
	stats     Stats
}

// MemoryOptions configures a memory cache
type MemoryOptions struct {
	MaxSize   int
	ShardSize int
}

// memoryShard represents a single shard of the cache
type memoryShard struct {
	items     map[Key]*list.Element
	lruList   *list.List
	maxSize   int
	mu        sync.RWMutex
	itemCount int
}

// cacheItem represents a cached item with TTL
type cacheItem struct {
	key       Key
	value     interface{}
	expiry    time.Time
	hasExpiry bool
}

// NewMemoryWithOptions creates a new memory cache with options
func NewMemoryWithOptions(opts MemoryOptions) *MemoryCache {
	// Default values
	if opts.MaxSize <= 0 {
		opts.MaxSize = 10000
	}
	
	if opts.ShardSize <= 0 {
		opts.ShardSize = 16
	}
	
	// Ensure ShardSize is a power of 2
	shardSize := nextPowerOfTwo(uint32(opts.ShardSize))
	shardMask := shardSize - 1
	
	// Calculate items per shard
	itemsPerShard := opts.MaxSize / int(shardSize)
	if itemsPerShard <= 0 {
		itemsPerShard = 100
	}
	
	// Create shards
	shards := make([]*memoryShard, shardSize)
	for i := uint32(0); i < shardSize; i++ {
		shards[i] = &memoryShard{
			items:   make(map[Key]*list.Element),
			lruList: list.New(),
			maxSize: itemsPerShard,
		}
	}
	
	cache := &MemoryCache{
		shards:    shards,
		shardMask: shardMask,
	}
	
	// Start cleanup worker
	go cache.cleanupWorker()
	
	return cache
}

// Get retrieves a value from the cache
func (c *MemoryCache) Get(key Key) (interface{}, bool) {
	shard := c.getShard(key)
	shard.mu.RLock()
	element, found := shard.items[key]
	
	if !found {
		shard.mu.RUnlock()
		atomic.AddUint64(&c.stats.Misses, 1)
		return nil, false
	}
	
	item := element.Value.(*cacheItem)
	
	// Check if expired
	if item.hasExpiry && time.Now().After(item.expiry) {
		shard.mu.RUnlock()
		// Delete in a separate goroutine to avoid deadlock
		go c.Delete(key)
		atomic.AddUint64(&c.stats.Misses, 1)
		atomic.AddUint64(&c.stats.Expirations, 1)
		return nil, false
	}
	
	shard.mu.RUnlock()
	
	// Move to front of LRU list (requires write lock)
	shard.mu.Lock()
	shard.lruList.MoveToFront(element)
	shard.mu.Unlock()
	
	atomic.AddUint64(&c.stats.Hits, 1)
	return item.value, true
}

// Set stores a value in the cache
func (c *MemoryCache) Set(key Key, value interface{}, ttl time.Duration) {
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	
	// Create cache item
	item := &cacheItem{
		key:   key,
		value: value,
	}
	
	// Set expiry if TTL provided
	if ttl > 0 {
		item.hasExpiry = true
		item.expiry = time.Now().Add(ttl)
	}
	
	// Check if key already exists
	if element, found := shard.items[key]; found {
		// Update existing item
		element.Value = item
		shard.lruList.MoveToFront(element)
		return
	}
	
	// Add new item
	element := shard.lruList.PushFront(item)
	shard.items[key] = element
	shard.itemCount++
	
	// Evict if needed
	c.evictIfNeeded(shard)
}

// Delete removes a value from the cache
func (c *MemoryCache) Delete(key Key) {
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	
	if element, found := shard.items[key]; found {
		c.removeElement(shard, element)
	}
}

// Clear removes all values from the cache
func (c *MemoryCache) Clear() {
	for _, shard := range c.shards {
		shard.mu.Lock()
		shard.items = make(map[Key]*list.Element)
		shard.lruList.Init()
		shard.itemCount = 0
		shard.mu.Unlock()
	}
	
	// Reset stats
	c.stats = Stats{}
}

// Size returns the number of items in the cache
func (c *MemoryCache) Size() int {
	var size int
	for _, shard := range c.shards {
		shard.mu.RLock()
		size += shard.itemCount
		shard.mu.RUnlock()
	}
	return size
}

// Stats returns cache statistics
func (c *MemoryCache) Stats() Stats {
	stats := Stats{
		Hits:        atomic.LoadUint64(&c.stats.Hits),
		Misses:      atomic.LoadUint64(&c.stats.Misses),
		Evictions:   atomic.LoadUint64(&c.stats.Evictions),
		Expirations: atomic.LoadUint64(&c.stats.Expirations),
		Size:        c.Size(),
	}
	return stats
}

// getShard returns the shard for a key
func (c *MemoryCache) getShard(key Key) *memoryShard {
	// Simple hash function for sharding
	hash := fnv32(string(key))
	return c.shards[hash&c.shardMask]
}

// evictIfNeeded evicts items if the shard is over capacity
func (c *MemoryCache) evictIfNeeded(shard *memoryShard) {
	for shard.itemCount > shard.maxSize {
		back := shard.lruList.Back()
		if back == nil {
			break
		}
		c.removeElement(shard, back)
		atomic.AddUint64(&c.stats.Evictions, 1)
	}
}

// removeElement removes an element from the cache
func (c *MemoryCache) removeElement(shard *memoryShard, element *list.Element) {
	item := element.Value.(*cacheItem)
	delete(shard.items, item.key)
	shard.lruList.Remove(element)
	shard.itemCount--
}

// cleanupWorker periodically removes expired items
func (c *MemoryCache) cleanupWorker() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		for _, shard := range c.shards {
			c.cleanupExpired(shard)
		}
	}
}

// cleanupExpired removes expired items from a shard
func (c *MemoryCache) cleanupExpired(shard *memoryShard) {
	now := time.Now()
	shard.mu.Lock()
	defer shard.mu.Unlock()
	
	var expiredItems []*list.Element
	
	// Find expired items
	for element := shard.lruList.Back(); element != nil; element = element.Prev() {
		item := element.Value.(*cacheItem)
		if item.hasExpiry && now.After(item.expiry) {
			expiredItems = append(expiredItems, element)
		} else {
			// LRU list is ordered, so once we hit a non-expired item, we can stop
			break
		}
	}
	
	// Remove expired items
	for _, element := range expiredItems {
		c.removeElement(shard, element)
		atomic.AddUint64(&c.stats.Expirations, 1)
	}
}

// nextPowerOfTwo returns the next power of two greater than or equal to x
func nextPowerOfTwo(x uint32) uint32 {
	if x == 0 {
		return 1
	}
	
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	
	return x + 1
}

// fnv32 implements a simple hash function
func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}