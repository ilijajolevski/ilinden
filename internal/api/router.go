// API routes definition
//
// Management API routing:
// - Route definitions
// - Handler mapping
// - Version management
// - Authentication requirements

package api

import (
	"net/http"
)

// Router manages API routes
type Router struct {
	mux *http.ServeMux
}

// NewRouter creates a new API router
func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

// Handler returns the HTTP handler for the router
func (r *Router) Handler() http.Handler {
	return r.mux
}

// RegisterHealthCheck registers a health check endpoint
func (r *Router) RegisterHealthCheck() {
	r.mux.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		WriteResponse(w, http.StatusOK, NewResponse(true, "OK", nil))
	})
}

// RegisterStatsEndpoint registers a stats endpoint
func (r *Router) RegisterStatsEndpoint(stats func() map[string]interface{}) {
	r.mux.HandleFunc("/stats", func(w http.ResponseWriter, req *http.Request) {
		WriteJSON(w, http.StatusOK, stats())
	})
}

// RegisterMetricsEndpoint registers a metrics endpoint
func (r *Router) RegisterMetricsEndpoint(metrics func() map[string]interface{}) {
	r.mux.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		WriteJSON(w, http.StatusOK, metrics())
	})
}

// RegisterVersionEndpoint registers a version endpoint
func (r *Router) RegisterVersionEndpoint(version, buildTime, gitCommit string) {
	r.mux.HandleFunc("/version", func(w http.ResponseWriter, req *http.Request) {
		info := map[string]string{
			"version":   version,
			"buildTime": buildTime,
			"gitCommit": gitCommit,
		}
		WriteJSON(w, http.StatusOK, info)
	})
}