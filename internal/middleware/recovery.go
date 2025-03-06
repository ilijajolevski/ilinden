// Panic recovery middleware
//
// Prevents server crashes from panics:
// - Panic catching
// - Error logging
// - Client error responses
// - Stack trace capture

package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/ilijajolevski/ilinden/internal/api"
	"github.com/ilijajolevski/ilinden/internal/telemetry"
)

// Recovery returns a middleware that recovers from panics
func Recovery(logger telemetry.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the error and stack trace
					stack := debug.Stack()
					logger.Error("Panic recovered",
						"error", fmt.Sprintf("%v", err),
						"stack", string(stack),
						"path", r.URL.Path,
						"method", r.Method,
					)
					
					// Return a 500 error to the client
					apiErr := api.NewError("Internal server error", "panic", http.StatusInternalServerError)
					api.WriteError(w, apiErr)
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}