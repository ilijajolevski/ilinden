// TTL management strategies
//
// Dynamic TTL calculation:
// - Content-type based TTL
// - Playlist type detection
// - Adaptive TTL for changing content
// - Jitter to prevent stampedes

package cache

import (
	"crypto/rand"
	"math"
	"net/http"
	"strings"
	"time"
)

// TTLStrategy represents a function that determines cache TTL
type TTLStrategy func(r *http.Request, resp *http.Response) time.Duration

// TTLOptions configures TTL calculation
type TTLOptions struct {
	DefaultTTL  time.Duration
	MasterTTL   time.Duration
	MediaTTL    time.Duration
	ApplyJitter bool
	JitterPct   float64 // Percentage of jitter (0-1)
}

// DefaultTTLOptions returns sensible default TTL options
func DefaultTTLOptions() TTLOptions {
	return TTLOptions{
		DefaultTTL:  10 * time.Second,
		MasterTTL:   30 * time.Second,
		MediaTTL:    5 * time.Second,
		ApplyJitter: true,
		JitterPct:   0.2, // 20% jitter
	}
}

// NewHLSTTLStrategy creates a TTL strategy for HLS content
func NewHLSTTLStrategy(opts TTLOptions) TTLStrategy {
	return func(r *http.Request, resp *http.Response) time.Duration {
		// Start with the default TTL
		ttl := opts.DefaultTTL
		
		// Check content type for specific handling
		contentType := resp.Header.Get("Content-Type")
		
		// HLS-specific TTL
		switch {
		case strings.Contains(contentType, "application/vnd.apple.mpegurl"),
		     strings.Contains(contentType, "application/x-mpegurl"):
			// Determine if master or media playlist
			if isMasterPlaylist(r, resp) {
				ttl = opts.MasterTTL
			} else {
				ttl = opts.MediaTTL
			}
		}
		
		// Apply jitter if enabled
		if opts.ApplyJitter && opts.JitterPct > 0 {
			ttl = applyJitter(ttl, opts.JitterPct)
		}
		
		return ttl
	}
}

// isMasterPlaylist attempts to determine if a response is a master playlist
func isMasterPlaylist(r *http.Request, resp *http.Response) bool {
	// Check URL path for common indicators
	path := r.URL.Path
	
	// Paths containing these terms are often master playlists
	if strings.Contains(path, "master") || 
	   strings.Contains(path, "variant") || 
	   strings.Contains(path, "playlist") {
		return true
	}
	
	// Paths containing these terms are often media playlists
	if strings.Contains(path, "media") || 
	   strings.Contains(path, "chunklist") || 
	   strings.Contains(path, "segment") {
		return false
	}
	
	// Default to shorter TTL (media playlist) if uncertain
	return false
}

// applyJitter adds random jitter to a TTL to prevent cache stampedes
func applyJitter(ttl time.Duration, jitterPct float64) time.Duration {
	if jitterPct <= 0 {
		return ttl
	}
	
	// Clamp jitter to 0-1 range
	if jitterPct > 1.0 {
		jitterPct = 1.0
	}
	
	// Calculate jitter range
	jitterRange := float64(ttl) * jitterPct
	
	// Generate random jitter value between -jitterRange/2 and +jitterRange/2
	jitterValue := (randomFloat() - 0.5) * jitterRange
	
	// Apply jitter
	newTTL := time.Duration(float64(ttl) + jitterValue)
	
	// Ensure TTL doesn't go below 1ms
	if newTTL < time.Millisecond {
		newTTL = time.Millisecond
	}
	
	return newTTL
}

// randomFloat returns a random float64 in the range [0.0, 1.0)
func randomFloat() float64 {
	var buf [8]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		// If random fails, use a fixed value
		return 0.5
	}
	
	// Convert to float between 0 and 1
	// IEEE 754 doubles have 52 bits of mantissa, so we're using the first 52 bits of the 64
	val := uint64(buf[0])<<56 | uint64(buf[1])<<48 | uint64(buf[2])<<40 | 
	       uint64(buf[3])<<32 | uint64(buf[4])<<24 | uint64(buf[5])<<16 | 
	       uint64(buf[6])<<8 | uint64(buf[7])
	
	// Scale to [0, 1)
	return float64(val) / float64(math.MaxUint64)
}