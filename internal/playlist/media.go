// Media playlist specific handling
//
// Media playlist (chunklist) specific logic:
// - Segment URL rewriting
// - Media sequence handling
// - Live window tracking
// - Discontinuity handling

package playlist

import (
	"net/url"

	"github.com/ilijajolevski/ilinden/pkg/hls"
)

// MediaProcessor handles media playlist processing
type MediaProcessor struct {
	baseURL  *url.URL
	proxyURL *url.URL
	options  ProcessorOptions
}

// NewMediaProcessor creates a new media playlist processor
func NewMediaProcessor(baseURL, proxyURL *url.URL, options ProcessorOptions) *MediaProcessor {
	return &MediaProcessor{
		baseURL:  baseURL,
		proxyURL: proxyURL,
		options:  options,
	}
}

// Process processes a media playlist
func (p *MediaProcessor) Process(playlist *hls.Playlist, token string) error {
	if !playlist.IsMedia() {
		return ErrNotMediaPlaylist
	}
	
	// Process each segment in the media playlist
	for i := range playlist.Media.Segments {
		if err := p.processSegment(&playlist.Media.Segments[i], token); err != nil {
			return err
		}
	}
	
	return nil
}

// processSegment processes a segment in a media playlist
func (p *MediaProcessor) processSegment(segment *hls.Segment, token string) error {
	// Skip empty URIs
	if segment.URI == "" {
		return nil
	}
	
	// Resolve URI to absolute URL if it's relative
	resolvedURL, err := resolveURL(p.baseURL, segment.URI)
	if err != nil {
		return err
	}
	
	// For segments, point directly to origin with token
	directURL := p.addTokenToURL(resolvedURL, token)
	segment.URI = directURL
	
	// Process key if present
	if segment.Key != nil {
		if err := p.processKey(segment.Key, token); err != nil {
			return err
		}
	}
	
	// Process map if present
	if segment.Map != nil {
		if err := p.processMap(segment.Map, token); err != nil {
			return err
		}
	}
	
	return nil
}

// processKey processes a segment key
func (p *MediaProcessor) processKey(key *hls.Key, token string) error {
	// Skip empty URIs
	if key.URI == "" {
		return nil
	}
	
	// Resolve URI to absolute URL if it's relative
	resolvedURL, err := resolveURL(p.baseURL, key.URI)
	if err != nil {
		return err
	}
	
	// Point directly to origin with token
	directURL := p.addTokenToURL(resolvedURL, token)
	key.URI = directURL
	
	return nil
}

// processMap processes a segment map
func (p *MediaProcessor) processMap(m *hls.Map, token string) error {
	// Skip empty URIs
	if m.URI == "" {
		return nil
	}
	
	// Resolve URI to absolute URL if it's relative
	resolvedURL, err := resolveURL(p.baseURL, m.URI)
	if err != nil {
		return err
	}
	
	// Point directly to origin with token
	directURL := p.addTokenToURL(resolvedURL, token)
	m.URI = directURL
	
	return nil
}

// addTokenToURL adds a token to a URL
func (p *MediaProcessor) addTokenToURL(targetURL *url.URL, token string) string {
	// Skip if no token or no token param name
	if token == "" || p.options.TokenParamName == "" {
		return targetURL.String()
	}
	
	// Clone the URL to avoid modifying the original
	result := *targetURL
	
	// Add token to query string
	q := result.Query()
	q.Set(p.options.TokenParamName, token)
	result.RawQuery = q.Encode()
	
	return result.String()
}