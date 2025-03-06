// Playlist URL modification
//
// Handles modification of URLs in playlists:
// - JWT token insertion
// - Relative to absolute URL conversion
// - Base URL handling
// - URL validation and encoding

package playlist

import (
	"errors"
	"net/url"
	"strings"

	"github.com/ilijajolevski/ilinden/pkg/hls"
)

// Common errors
var (
	ErrInvalidBaseURL      = errors.New("invalid base URL")
	ErrInvalidProxyURL     = errors.New("invalid proxy URL")
	ErrInvalidPlaylist     = errors.New("invalid playlist")
	ErrNotMasterPlaylist   = errors.New("not a master playlist")
	ErrNotMediaPlaylist    = errors.New("not a media playlist")
	ErrEmptyToken          = errors.New("empty token")
	ErrEmptyTokenParamName = errors.New("empty token parameter name")
)

// ProcessorOptions configures the playlist processor
type ProcessorOptions struct {
	TokenParamName string // Query parameter name for the token
	PathParamName  string // Parameter name for the path in the proxy URL
	UsePathParam   bool   // Whether to use the path parameter for the target URL
}

// DefaultProcessorOptions returns the default processor options
func DefaultProcessorOptions() ProcessorOptions {
	return ProcessorOptions{
		TokenParamName: "token",
		PathParamName:  "url",
		UsePathParam:   false,
	}
}

// Modifier handles playlist URL modification
type Modifier struct {
	options ProcessorOptions
}

// NewModifier creates a new playlist modifier
func NewModifier(options ProcessorOptions) *Modifier {
	return &Modifier{
		options: options,
	}
}

// Process processes a playlist by modifying its URLs
func (m *Modifier) Process(playlist *hls.Playlist, baseURL, proxyURL *url.URL, token string) error {
	// Basic validation
	if baseURL == nil {
		return ErrInvalidBaseURL
	}
	
	if proxyURL == nil {
		return ErrInvalidProxyURL
	}
	
	if playlist == nil {
		return ErrInvalidPlaylist
	}
	
	if token == "" {
		return ErrEmptyToken
	}
	
	if m.options.TokenParamName == "" {
		return ErrEmptyTokenParamName
	}
	
	// Process according to playlist type
	switch playlist.Type {
	case hls.PlaylistTypeMaster:
		processor := NewMasterProcessor(baseURL, proxyURL, m.options)
		return processor.Process(playlist, token)
		
	case hls.PlaylistTypeMedia:
		processor := NewMediaProcessor(baseURL, proxyURL, m.options)
		return processor.Process(playlist, token)
		
	default:
		return ErrInvalidPlaylist
	}
}

// resolveURL resolves a URL that may be relative to a base URL
func resolveURL(baseURL *url.URL, urlStr string) (*url.URL, error) {
	// Skip empty URLs
	if urlStr == "" {
		return nil, errors.New("empty URL")
	}
	
	// Check if the URL is already absolute
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	
	// If it's already absolute, return it
	if parsedURL.IsAbs() {
		return parsedURL, nil
	}
	
	// Otherwise, resolve it against the base URL
	return baseURL.ResolveReference(parsedURL), nil
}

// IsM3U8 checks if a URL is likely an M3U8 playlist
func IsM3U8(urlStr string) bool {
	return strings.HasSuffix(strings.ToLower(urlStr), ".m3u8")
}