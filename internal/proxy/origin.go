// Origin server communication
//
// Handles all interaction with origin servers:
// - Request formation
// - Header manipulation
// - Response handling
// - Error mapping
// - Retry logic

package proxy

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ilijajolevski/ilinden/internal/config"
	"github.com/ilijajolevski/ilinden/internal/telemetry"
)

// OriginHandler manages communication with origin servers
type OriginHandler struct {
	client  *http.Client
	config  *config.OriginConfig
	metrics telemetry.Metrics
	logger  telemetry.Logger
}

// OriginRequest represents a request to the origin server
type OriginRequest struct {
	Method  string
	URL     *url.URL
	Headers http.Header
	Body    io.Reader
}

// NewOriginHandler creates a new origin handler
func NewOriginHandler(config *config.OriginConfig, metrics telemetry.Metrics, logger telemetry.Logger) *OriginHandler {
	// Create transport with connection pooling
	transport := &http.Transport{
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
	}

	// Create client with timeout
	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return &OriginHandler{
		client:  client,
		config:  config,
		metrics: metrics,
		logger:  logger,
	}
}

// Do sends a request to the origin server
func (h *OriginHandler) Do(ctx context.Context, req *OriginRequest) (*http.Response, error) {
	// Start timing
	startTime := time.Now()
	
	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL.String(), req.Body)
	if err != nil {
		return nil, err
	}
	
	// Copy headers
	for k, vv := range req.Headers {
		for _, v := range vv {
			httpReq.Header.Add(k, v)
		}
	}
	
	// Send request to origin
	resp, err := h.client.Do(httpReq)
	
	// Record metrics
	h.metrics.ObserveOriginDuration(req.URL.Host, time.Since(startTime))
	
	// Handle errors
	if err != nil {
		h.metrics.IncCounter("origin.error")
		h.logger.Error("Origin request failed", "error", err.Error(), "url", req.URL.String())
		return nil, h.mapError(err)
	}
	
	// Record status code metrics
	h.metrics.IncCounter("origin.status." + http.StatusText(resp.StatusCode))
	
	return resp, nil
}

// GetURL constructs a URL for the origin server
func (h *OriginHandler) GetURL(path string) (*url.URL, error) {
	baseURL, err := url.Parse(h.config.BaseURL)
	if err != nil {
		return nil, err
	}
	
	// Check if path is already a full URL
	if pathURL, err := url.Parse(path); err == nil && pathURL.IsAbs() {
		return pathURL, nil
	}
	
	// Use base scheme if not specified
	if baseURL.Scheme == "" {
		baseURL.Scheme = h.config.DefaultScheme
	}
	
	// Combine with path
	return baseURL.ResolveReference(&url.URL{Path: path}), nil
}

// mapError maps Go errors to proxy errors
func (h *OriginHandler) mapError(err error) error {
	// Check for timeout
	if err.Error() == "net/http: timeout awaiting response headers" {
		return ErrOriginTimeout
	}
	
	// Check for connection refused
	if err.Error() == "dial tcp: connect: connection refused" {
		return ErrOriginRefused
	}
	
	// Default to origin error
	return &ProxyError{
		Code:    http.StatusBadGateway,
		Message: "Origin error",
		Err:     err,
	}
}