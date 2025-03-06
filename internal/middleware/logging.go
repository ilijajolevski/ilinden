// Request logging middleware
//
// HTTP request/response logging:
// - Structured logging
// - Timing information
// - Success/failure recording
// - Sampling for high-volume paths

package middleware

import (
	"net/http"
	"time"

	"github.com/ilijajolevski/ilinden/internal/telemetry"
)

// responseWriter is a wrapper for http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response size
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Status returns the status code
func (rw *responseWriter) Status() int {
	if rw.statusCode == 0 {
		return http.StatusOK
	}
	return rw.statusCode
}

// Size returns the response size
func (rw *responseWriter) Size() int {
	return rw.size
}

// Logging returns a middleware that logs requests
func Logging(logger telemetry.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Create a wrapper for the response writer
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     0,
				size:           0,
			}
			
			// Call the next handler
			next.ServeHTTP(rw, r)
			
			// Calculate duration
			duration := time.Since(start)
			
			// Log the request
			logger.Info("Request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.Status(),
				"duration", duration.String(),
				"size", rw.Size(),
				"remote", r.RemoteAddr,
				"user-agent", r.UserAgent(),
			)
		})
	}
}