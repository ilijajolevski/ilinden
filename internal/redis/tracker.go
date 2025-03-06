// Player tracking implementation
//
// Redis-based player tracking:
// - Activity recording
// - Session tracking
// - Analytics support
// - Efficient data structures

package redis

import (
	"sync"
	"time"

	"github.com/ilijajolevski/ilinden/internal/config"
	"github.com/ilijajolevski/ilinden/internal/telemetry"
)

// Tracker handles player activity tracking
type Tracker struct {
	config *config.RedisConfig
	logger telemetry.Logger

	// For this simple implementation, we'll use an in-memory map
	// In a real implementation, this would use Redis
	players     map[string]*PlayerInfo
	mu          sync.RWMutex
	trackExpiry time.Duration
}

// PlayerInfo represents player tracking information
type PlayerInfo struct {
	PlayerID       string
	LastActivity   time.Time
	Path           string
	UserAgent      string
	FirstSeen      time.Time
	ActivityCount  int
}

// NewTracker creates a new player tracker
func NewTracker(config *config.RedisConfig, logger telemetry.Logger) *Tracker {
	return &Tracker{
		config:      config,
		logger:      logger,
		players:     make(map[string]*PlayerInfo),
		trackExpiry: config.TrackingExpiry,
	}
}

// TrackPlayer tracks player activity
func (t *Tracker) TrackPlayer(playerID, path, userAgent string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()

	// Check if player exists
	player, exists := t.players[playerID]
	if !exists {
		// Create new player
		player = &PlayerInfo{
			PlayerID:      playerID,
			LastActivity:  now,
			Path:          path,
			UserAgent:     userAgent,
			FirstSeen:     now,
			ActivityCount: 1,
		}
		t.players[playerID] = player
	} else {
		// Update existing player
		player.LastActivity = now
		player.Path = path
		player.ActivityCount++
	}

	// In a real implementation, this would be sent to Redis
	// with proper TTL expiration
}

// GetActivePlayers returns the number of active players
func (t *Tracker) GetActivePlayers() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// In a real implementation, this would query Redis
	// for the count of active players
	count := 0
	now := time.Now()
	cutoff := now.Add(-t.trackExpiry)

	for _, player := range t.players {
		if player.LastActivity.After(cutoff) {
			count++
		}
	}

	return count
}

// GetPlayerInfo returns information about a player
func (t *Tracker) GetPlayerInfo(playerID string) *PlayerInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// In a real implementation, this would query Redis
	player, exists := t.players[playerID]
	if !exists {
		return nil
	}

	return player
}

// StartCleanupWorker starts a worker to clean up expired players
func (t *Tracker) StartCleanupWorker() {
	// In a real implementation, Redis TTL would handle this automatically
	// This is just a simple in-memory implementation
	ticker := time.NewTicker(t.trackExpiry / 2)
	go func() {
		for range ticker.C {
			t.cleanup()
		}
	}()
}

// cleanup removes expired players
func (t *Tracker) cleanup() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-t.trackExpiry)

	for id, player := range t.players {
		if player.LastActivity.Before(cutoff) {
			delete(t.players, id)
		}
	}
}