// API request handlers
//
// Admin API endpoints:
// - Status information
// - Cache management
// - Configuration reporting
// - Health checks
// - Player statistics

package api

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

// StatusHandler returns a handler for the /status endpoint
func StatusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"uptime":    time.Since(startTime).String(),
			"go_version": runtime.Version(),
			"goroutines": runtime.NumGoroutine(),
		}
		
		WriteJSON(w, http.StatusOK, stats)
	}
}

// HealthHandler returns a handler for the /health endpoint
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := map[string]interface{}{
			"status": "ok",
		}
		
		WriteJSON(w, http.StatusOK, health)
	}
}

// ConfigHandler returns a handler for the /config endpoint
func ConfigHandler(configGetter func() interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config := configGetter()
		WriteJSON(w, http.StatusOK, config)
	}
}

// CacheStatsHandler returns a handler for the /cache/stats endpoint
func CacheStatsHandler(statsGetter func() interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := statsGetter()
		WriteJSON(w, http.StatusOK, stats)
	}
}

// CacheClearHandler returns a handler for the /cache/clear endpoint
func CacheClearHandler(clearFunc func() error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			WriteError(w, NewError("Method not allowed", "method_not_allowed", http.StatusMethodNotAllowed))
			return
		}
		
		err := clearFunc()
		if err != nil {
			WriteError(w, NewError("Failed to clear cache", "clear_failed", http.StatusInternalServerError))
			return
		}
		
		WriteResponse(w, http.StatusOK, NewResponse(true, "Cache cleared", nil))
	}
}

// PlayersHandler returns a handler for the /players endpoint
func PlayersHandler(playersGetter func() interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		players := playersGetter()
		WriteJSON(w, http.StatusOK, players)
	}
}

var startTime = time.Now()