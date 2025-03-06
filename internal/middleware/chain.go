// Middleware chaining utility
//
// Support for HTTP middleware chains:
// - Middleware composition
// - Order management
// - Context propagation
// - Chain building helpers

package middleware

import (
	"net/http"
)

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain represents a chain of middleware
type Chain struct {
	middlewares []Middleware
}

// NewChain creates a new middleware chain
func NewChain(middlewares ...Middleware) Chain {
	return Chain{
		middlewares: append([]Middleware{}, middlewares...),
	}
}

// Then applies the middleware chain to a handler
func (c Chain) Then(h http.Handler) http.Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}

// ThenFunc applies the middleware chain to a handler function
func (c Chain) ThenFunc(fn func(http.ResponseWriter, *http.Request)) http.Handler {
	return c.Then(http.HandlerFunc(fn))
}

// Append adds middleware to the chain
func (c Chain) Append(middlewares ...Middleware) Chain {
	newMiddlewares := make([]Middleware, len(c.middlewares)+len(middlewares))
	copy(newMiddlewares, c.middlewares)
	copy(newMiddlewares[len(c.middlewares):], middlewares)
	
	return Chain{
		middlewares: newMiddlewares,
	}
}

// Extend extends the chain with another chain
func (c Chain) Extend(chain Chain) Chain {
	return c.Append(chain.middlewares...)
}