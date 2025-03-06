// Main proxy request handler
//
// Core request processing logic:
// - Request path analysis
// - Playlist vs segment request detection
// - Appropriate handler dispatch
// - Error handling

package proxy

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ilijajolevski/ilinden/internal/api"
	"github.com/ilijajolevski/ilinden/internal/cache"
	"github.com/ilijajolevski/ilinden/internal/config"
	"github.com/ilijajolevski/ilinden/internal/jwt"
	"github.com/ilijajolevski/ilinden/internal/playlist"
	"github.com/ilijajolevski/ilinden/internal/redis"
	"github.com/ilijajolevski/ilinden/internal/telemetry"
)

// Common errors
var (
	ErrNoTargetURL      = errors.New("no target URL provided")
	ErrInvalidTargetURL = errors.New("invalid target URL")
	ErrOriginError      = errors.New("origin server error")
	ErrParsingPlaylist  = errors.New("error parsing playlist")
)

// Handler handles proxy requests
type Handler struct {
	config         *config.Config
	jwtExtractor   *jwt.Extractor
	jwtValidator   *jwt.Validator
	cache          cache.Cache
	logger         telemetry.Logger
	metrics        telemetry.Metrics
	playlistParser *playlist.Parser
	redisTracker   *redis.Tracker
	originClient   *http.Client
}

// HandlerOptions contains options for creating a new handler
type HandlerOptions struct {
	Config       *config.Config
	Cache        cache.Cache
	Logger       telemetry.Logger
	Metrics      telemetry.Metrics
	RedisTracker *redis.Tracker
}

// NewHandler creates a new proxy handler
func NewHandler(opts HandlerOptions) *Handler {
	// Create origin client
	originClient := &http.Client{
		Timeout: opts.Config.Origin.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:          opts.Config.Origin.MaxIdleConns,
			MaxIdleConnsPerHost:   opts.Config.Origin.MaxIdleConnsPerHost,
			MaxConnsPerHost:       opts.Config.Origin.MaxConnsPerHost,
			IdleConnTimeout:       opts.Config.Origin.IdleConnTimeout,
			TLSHandshakeTimeout:   opts.Config.Origin.TLSHandshakeTimeout,
			ExpectContinueTimeout: opts.Config.Origin.ExpectContinueTimeout,
		},
	}

	// Create JWT components
	jwtExtractor := jwt.NewExtractor(&opts.Config.JWT)
	jwtValidator := jwt.NewValidator(&opts.Config.JWT, opts.Cache)

	return &Handler{
		config:         opts.Config,
		jwtExtractor:   jwtExtractor,
		jwtValidator:   jwtValidator,
		cache:          opts.Cache,
		logger:         opts.Logger,
		metrics:        opts.Metrics,
		playlistParser: playlist.NewParser(),
		redisTracker:   opts.RedisTracker,
		originClient:   originClient,
	}
}

// ServeHTTP handles HTTP requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Start timing
	startTime := time.Now()
	
	// Extract token
	token, err := h.jwtExtractor.Extract(r)
	if err != nil {
		h.handleError(w, r, err, http.StatusUnauthorized)
		return
	}
	
	// Validate token
	claims, err := h.jwtValidator.ValidateToken(token)
	if err != nil {
		h.handleError(w, r, err, http.StatusUnauthorized)
		return
	}
	
	// Get player ID for tracking
	playerID, err := claims.GetPlayerID()
	if err != nil {
		h.logger.Warn("Failed to get player ID from token", "error", err.Error())
		// Continue without player ID
	}
	
	// Track player if tracking is enabled
	if h.redisTracker != nil && playerID != "" {
		h.redisTracker.TrackPlayer(playerID, r.URL.Path, r.Header.Get("User-Agent"))
	}
	
	// Determine target URL
	targetURL, err := h.getTargetURL(r)
	if err != nil {
		h.handleError(w, r, err, http.StatusBadRequest)
		return
	}
	
	// Check if the target is an HLS playlist
	isM3U8 := playlist.IsM3U8(targetURL.Path)
	
	// Set cache key based on URL and token
	keyPrefix := "playlist:"
	if isM3U8 {
		keyPrefix = "playlist:"
	} else {
		keyPrefix = "segment:"
	}
	cacheKey := cache.Key(keyPrefix + targetURL.String() + ":" + token)
	
	// Check cache first
	if h.config.Cache.Enabled {
		cachedContent, found := h.cache.Get(cacheKey)
		if found {
			if cachedBytes, ok := cachedContent.([]byte); ok {
				h.metrics.IncCounter("cache.hit")
				contentType := "application/octet-stream"
				if isM3U8 {
					contentType = "application/vnd.apple.mpegurl"
				}
				
				w.Header().Set("Content-Type", contentType)
				w.Header().Set("Content-Length", strconv.Itoa(len(cachedBytes)))
				w.Header().Set("X-Cache", "HIT")
				w.Write(cachedBytes)
				
				// Record metrics
				h.metrics.ObserveRequestDuration(r.URL.Path, time.Since(startTime))
				return
			}
		}
		h.metrics.IncCounter("cache.miss")
	}
	
	// Create request to origin
	originReq, err := http.NewRequestWithContext(r.Context(), "GET", targetURL.String(), nil)
	if err != nil {
		h.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	
	// Copy relevant headers from original request
	h.copyHeaders(r.Header, originReq.Header)
	
	// Send request to origin
	originResp, err := h.originClient.Do(originReq)
	if err != nil {
		h.handleError(w, r, err, http.StatusBadGateway)
		return
	}
	
	// Check if origin returned an error
	if originResp.StatusCode >= 400 {
		h.handleError(w, r, ErrOriginError, originResp.StatusCode)
		return
	}
	
	// Process the response
	if isM3U8 {
		// For M3U8 playlists, we need to process the content
		h.handlePlaylist(w, r, originResp, targetURL, token, cacheKey)
	} else {
		// For other content, just proxy the response
		h.handleRawContent(w, r, originResp, cacheKey)
	}
	
	// Record metrics
	h.metrics.ObserveRequestDuration(r.URL.Path, time.Since(startTime))
}

// handlePlaylist processes an HLS playlist
func (h *Handler) handlePlaylist(w http.ResponseWriter, r *http.Request, originResp *http.Response, targetURL *url.URL, token string, cacheKey cache.Key) {
	// Get processor options
	procOptions := playlist.ProcessorOptions{
		TokenParamName: h.config.JWT.ParamName,
		PathParamName:  "url",
		UsePathParam:   false,
	}
	
	// Create a proxy URL based on the current request
	proxyURL := &url.URL{
		Scheme: r.URL.Scheme,
		Host:   r.URL.Host,
		Path:   r.URL.Path,
	}
	
	// Process the playlist
	processedContent, err := h.playlistParser.ParseAndProcessResponse(
		originResp.Body,
		targetURL,
		proxyURL,
		token,
		procOptions,
	)
	
	if err != nil {
		h.handleError(w, r, fmt.Errorf("%w: %v", ErrParsingPlaylist, err), http.StatusInternalServerError)
		return
	}
	
	// Set appropriate headers
	contentType := originResp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/vnd.apple.mpegurl"
	}
	
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(processedContent)))
	w.Header().Set("X-Cache", "MISS")
	
	// Copy other relevant headers
	h.copyHeadersToResponse(originResp.Header, w.Header())
	
	// Cache the processed content if caching is enabled
	if h.config.Cache.Enabled {
		// Determine TTL based on playlist type
		var ttl time.Duration
		if strings.Contains(string(processedContent), "#EXT-X-STREAM-INF") {
			ttl = h.config.Cache.TTLMaster
		} else {
			ttl = h.config.Cache.TTLMedia
		}
		
		h.cache.Set(cacheKey, processedContent, ttl)
	}
	
	// Write the response
	w.Write(processedContent)
}

// handleRawContent proxies raw content without modification
func (h *Handler) handleRawContent(w http.ResponseWriter, r *http.Request, originResp *http.Response, cacheKey cache.Key) {
	// Set appropriate headers
	w.Header().Set("Content-Type", originResp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", originResp.Header.Get("Content-Length"))
	w.Header().Set("X-Cache", "MISS")
	
	// Copy other relevant headers
	h.copyHeadersToResponse(originResp.Header, w.Header())
	
	// Read and write the response body
	contentBytes, err := io.ReadAll(originResp.Body)
	if err != nil {
		h.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	
	// Cache the content if caching is enabled
	if h.config.Cache.Enabled {
		// Use a shorter TTL for segments
		h.cache.Set(cacheKey, contentBytes, h.config.Cache.TTLMedia)
	}
	
	// Write the response
	w.Write(contentBytes)
}

// getTargetURL extracts the target URL from the request
func (h *Handler) getTargetURL(r *http.Request) (*url.URL, error) {
	// Check if target URL is provided as a query parameter
	targetStr := r.URL.Query().Get("url")
	if targetStr != "" {
		targetURL, err := url.Parse(targetStr)
		if err != nil {
			return nil, ErrInvalidTargetURL
		}
		return targetURL, nil
	}
	
	// Otherwise, use the request path with the origin base URL
	originBaseURL := h.config.Origin.BaseURL
	if originBaseURL == "" {
		// If no base URL is configured, we cannot determine the target
		return nil, ErrNoTargetURL
	}
	
	// Parse origin base URL
	baseURL, err := url.Parse(originBaseURL)
	if err != nil {
		return nil, ErrInvalidTargetURL
	}
	
	// Combine with request path
	return baseURL.ResolveReference(&url.URL{Path: r.URL.Path, RawQuery: r.URL.RawQuery}), nil
}

// handleError handles errors in a consistent way
func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	// Log the error
	h.logger.Error("Proxy error", "error", err.Error(), "path", r.URL.Path, "status", statusCode)
	
	// Increment error metric
	h.metrics.IncCounter("error." + strconv.Itoa(statusCode))
	
	// JWT-specific errors
	var tokenErr *jwt.TokenError
	if errors.As(err, &tokenErr) {
		// Use the status code from the token error
		statusCode = tokenErr.StatusCode
		
		// Create API error response
		apiErr := api.NewError(tokenErr.Error(), "token_error", statusCode)
		api.WriteError(w, apiErr)
		return
	}
	
	// Generic error response
	message := "Internal server error"
	if statusCode == http.StatusBadRequest {
		message = "Bad request"
	} else if statusCode == http.StatusUnauthorized {
		message = "Unauthorized"
	} else if statusCode == http.StatusForbidden {
		message = "Forbidden"
	} else if statusCode == http.StatusNotFound {
		message = "Not found"
	} else if statusCode == http.StatusBadGateway {
		message = "Origin server error"
	}
	
	apiErr := api.NewError(message, "proxy_error", statusCode)
	api.WriteError(w, apiErr)
}

// copyHeaders copies headers from src to dst
func (h *Handler) copyHeaders(src, dst http.Header) {
	for k, vv := range src {
		// Skip certain headers
		if strings.HasPrefix(strings.ToLower(k), "x-") {
			continue
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// copyHeadersToResponse copies headers from origin response to client response
func (h *Handler) copyHeadersToResponse(src, dst http.Header) {
	for k, vv := range src {
		// Skip content headers that we set specifically
		if strings.ToLower(k) == "content-length" || strings.ToLower(k) == "content-type" {
			continue
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}