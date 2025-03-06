// HTTP client connection pooling
//
// Manages HTTP client connections to origin:
// - Persistent connection pooling
// - Connection reuse
// - Idle connection management
// - Connection health checking

package proxy

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ilijajolevski/ilinden/internal/config"
)

// ConnectionPool manages HTTP client connection pooling
type ConnectionPool struct {
	transport     *http.Transport
	originClients map[string]*http.Client
	config        *config.OriginConfig
	mu            sync.RWMutex
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config *config.OriginConfig) *ConnectionPool {
	// Create base transport
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   config.Timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
	}

	return &ConnectionPool{
		transport:     transport,
		originClients: make(map[string]*http.Client),
		config:        config,
	}
}

// GetClient returns a client for the given origin
func (p *ConnectionPool) GetClient(originHost string) *http.Client {
	p.mu.RLock()
	client, exists := p.originClients[originHost]
	p.mu.RUnlock()

	if exists {
		return client
	}

	// Create a new client if one doesn't exist
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check again in case another goroutine created the client
	if client, exists := p.originClients[originHost]; exists {
		return client
	}

	// Create a clone of the transport
	transportClone := p.transport.Clone()

	// Create a new client
	client = &http.Client{
		Transport: transportClone,
		Timeout:   p.config.Timeout,
	}

	p.originClients[originHost] = client
	return client
}

// CloseIdleConnections closes idle connections for all clients
func (p *ConnectionPool) CloseIdleConnections() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, client := range p.originClients {
		client.CloseIdleConnections()
	}
}

// GetDefaultClient returns a default client
func (p *ConnectionPool) GetDefaultClient() *http.Client {
	return &http.Client{
		Transport: p.transport.Clone(),
		Timeout:   p.config.Timeout,
	}
}