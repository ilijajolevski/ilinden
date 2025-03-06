// Metrics collection middleware
//
// Performance and operation metrics:
// - Request counters
// - Latency histograms
// - Status code tracking
// - Cache hit/miss recording

package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ilijajolevski/ilinden/internal/telemetry"
)

// Metrics returns a middleware that collects metrics
func Metrics(metrics telemetry.Metrics) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Create a wrapper for the response writer
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     0,
				size:           0,
			}
			
			// Increment request counter
			metrics.IncCounter("request.total")
			metrics.IncCounter("request.method." + r.Method)
			
			// Call the next handler
			next.ServeHTTP(rw, r)
			
			// Calculate duration
			duration := time.Since(start)
			
			// Record metrics
			metrics.IncCounter("response.status." + strconv.Itoa(rw.Status()))
			metrics.ObserveRequestDuration(r.URL.Path, duration)
			
			// Record size metrics
			sizeKB := float64(rw.Size()) / 1024.0
			metrics.ObserveHistogram("response.size", sizeKB)
		})
	}
}