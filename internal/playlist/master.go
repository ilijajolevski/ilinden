// Master playlist specific handling
//
// Special processing for master playlists:
// - BANDWIDTH, RESOLUTION tag preservation
// - Variant stream handling
// - Alternative stream handling

package playlist

import (
	"net/url"
	"strings"

	"github.com/ilijajolevski/ilinden/pkg/hls"
)

// MasterProcessor handles master playlist processing
type MasterProcessor struct {
	baseURL  *url.URL
	proxyURL *url.URL
	options  ProcessorOptions
}

// NewMasterProcessor creates a new master playlist processor
func NewMasterProcessor(baseURL, proxyURL *url.URL, options ProcessorOptions) *MasterProcessor {
	return &MasterProcessor{
		baseURL:  baseURL,
		proxyURL: proxyURL,
		options:  options,
	}
}

// Process processes a master playlist
func (p *MasterProcessor) Process(playlist *hls.Playlist, token string) error {
	if !playlist.IsMaster() {
		return ErrNotMasterPlaylist
	}
	
	// Process each variant stream in the master playlist
	for i := range playlist.Master.Variants {
		if err := p.processVariant(&playlist.Master.Variants[i], token); err != nil {
			return err
		}
	}
	
	// Process each I-frame stream
	for i := range playlist.Master.IFrameStreams {
		if err := p.processIFrameStream(&playlist.Master.IFrameStreams[i], token); err != nil {
			return err
		}
	}
	
	// Process each media group
	for _, mediaGroups := range playlist.Master.MediaGroups {
		for i := range mediaGroups {
			if err := p.processMediaGroup(&mediaGroups[i], token); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// processVariant processes a variant stream in a master playlist
func (p *MasterProcessor) processVariant(variant *hls.Variant, token string) error {
	// Skip empty URIs
	if variant.URI == "" {
		return nil
	}
	
	// Resolve URI to absolute URL if it's relative
	resolvedURL, err := resolveURL(p.baseURL, variant.URI)
	if err != nil {
		return err
	}
	
	// Point the variant back to our proxy with the token
	proxyPath := p.generateProxyPath(resolvedURL, token)
	variant.URI = proxyPath
	
	return nil
}

// processIFrameStream processes an I-frame stream in a master playlist
func (p *MasterProcessor) processIFrameStream(iframe *hls.IFrameStream, token string) error {
	// Skip empty URIs
	if iframe.URI == "" {
		return nil
	}
	
	// Resolve URI to absolute URL if it's relative
	resolvedURL, err := resolveURL(p.baseURL, iframe.URI)
	if err != nil {
		return err
	}
	
	// Point the I-frame stream back to our proxy with the token
	proxyPath := p.generateProxyPath(resolvedURL, token)
	iframe.URI = proxyPath
	
	return nil
}

// processMediaGroup processes a media group in a master playlist
func (p *MasterProcessor) processMediaGroup(media *hls.MediaGroup, token string) error {
	// Skip empty URIs
	if media.URI == "" {
		return nil
	}
	
	// Resolve URI to absolute URL if it's relative
	resolvedURL, err := resolveURL(p.baseURL, media.URI)
	if err != nil {
		return err
	}
	
	// Point the media group back to our proxy with the token
	proxyPath := p.generateProxyPath(resolvedURL, token)
	media.URI = proxyPath
	
	return nil
}

// generateProxyPath creates a proxy path for the variant
func (p *MasterProcessor) generateProxyPath(targetURL *url.URL, token string) string {
	// Use proxy host as base
	result := &url.URL{
		Path: p.proxyURL.Path,
	}
	
	// Add the token
	if p.options.TokenParamName != "" && token != "" {
		q := result.Query()
		q.Set(p.options.TokenParamName, token)
		result.RawQuery = q.Encode()
	}
	
	// Add target URL as path or in special parameter
	if p.options.UsePathParam {
		// Add target as a query parameter
		q := result.Query()
		q.Set(p.options.PathParamName, targetURL.String())
		result.RawQuery = q.Encode()
	} else {
		// Add target as part of the path
		newPath := strings.TrimSuffix(p.proxyURL.Path, "/")
		if !strings.HasPrefix(targetURL.Path, "/") {
			newPath += "/"
		}
		newPath += targetURL.Path
		
		// Add target query string
		result.Path = newPath
		if targetURL.RawQuery != "" {
			if result.RawQuery != "" {
				result.RawQuery += "&" + targetURL.RawQuery
			} else {
				result.RawQuery = targetURL.RawQuery
			}
		}
	}
	
	return result.String()
}