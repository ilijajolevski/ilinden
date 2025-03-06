// HLS playlist parsing
//
// Efficient parsing of HLS playlists:
// - Streaming parser design
// - Memory-efficient processing
// - HLS protocol compliance
// - Error handling and recovery

package playlist

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/ilijajolevski/ilinden/pkg/hls"
)

// Parser handles HLS playlist parsing
type Parser struct {
	hlsParser *hls.Parser
}

// NewParser creates a new HLS playlist parser
func NewParser() *Parser {
	return &Parser{
		hlsParser: hls.New(),
	}
}

// Parse parses an HLS playlist from a reader
func (p *Parser) Parse(r io.Reader) (*hls.Playlist, error) {
	return p.hlsParser.Parse(r)
}

// ParseAndProcess parses and processes a playlist
func (p *Parser) ParseAndProcess(r io.Reader, baseURL, proxyURL *url.URL, token string, options ProcessorOptions) (string, error) {
	// Parse the playlist
	playlist, err := p.Parse(r)
	if err != nil {
		return "", err
	}
	
	// Process the playlist
	modifier := NewModifier(options)
	if err := modifier.Process(playlist, baseURL, proxyURL, token); err != nil {
		return "", err
	}
	
	// Convert back to string
	return playlist.String(), nil
}

// ParseAndProcessBytes parses and processes a playlist from bytes
func (p *Parser) ParseAndProcessBytes(playlistData []byte, baseURL, proxyURL *url.URL, token string, options ProcessorOptions) ([]byte, error) {
	// Parse the playlist
	reader := bytes.NewReader(playlistData)
	playlist, err := p.Parse(reader)
	if err != nil {
		return nil, err
	}
	
	// Process the playlist
	modifier := NewModifier(options)
	if err := modifier.Process(playlist, baseURL, proxyURL, token); err != nil {
		return nil, err
	}
	
	// Convert back to bytes
	return []byte(playlist.String()), nil
}

// ParseAndProcessResponse parses and processes a playlist from an HTTP response
func (p *Parser) ParseAndProcessResponse(body io.ReadCloser, baseURL, proxyURL *url.URL, token string, options ProcessorOptions) ([]byte, error) {
	// Read the entire body
	defer body.Close()
	
	playlistData, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	
	// Parse and process
	return p.ParseAndProcessBytes(playlistData, baseURL, proxyURL, token, options)
}

// DetectPlaylistType attempts to determine the type of playlist based on content
func DetectPlaylistType(content []byte) hls.PlaylistType {
	contentStr := string(content)
	
	// Check for master playlist indicators
	if strings.Contains(contentStr, "#EXT-X-STREAM-INF") {
		return hls.PlaylistTypeMaster
	}
	
	// Check for media playlist indicators
	if strings.Contains(contentStr, "#EXTINF") ||
	   strings.Contains(contentStr, "#EXT-X-TARGETDURATION") {
		return hls.PlaylistTypeMedia
	}
	
	// Unknown or invalid
	return hls.PlaylistTypeUnknown
}