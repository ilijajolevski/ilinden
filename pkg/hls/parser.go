// Generic HLS parser
//
// Low-level HLS parsing:
// - Line-by-line processing
// - Tag identification
// - Attribute parsing
// - Protocol compliance checking

package hls

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// Common errors
var (
	ErrPlaylistFormat = errors.New("invalid playlist format")
	ErrPlaylistHeader = errors.New("missing #EXTM3U header")
	ErrTagFormat      = errors.New("invalid tag format")
)

// Parser represents an HLS playlist parser
type Parser struct {
	playlist *Playlist
}

// New creates a new HLS parser
func New() *Parser {
	return &Parser{
		playlist: NewPlaylist(),
	}
}

// Parse parses an HLS playlist from a reader
func (p *Parser) Parse(r io.Reader) (*Playlist, error) {
	scanner := bufio.NewScanner(r)
	lineNum := 0
	var lastTag *Tag
	var err error
	
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		
		// Store all raw lines
		p.playlist.RawLines = append(p.playlist.RawLines, line)
		
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		// First line must be #EXTM3U
		if lineNum == 1 {
			if line != TagExtM3U {
				return nil, ErrPlaylistHeader
			}
			p.playlist.OriginalHeader = line
			continue
		}
		
		// Handle tags
		if strings.HasPrefix(line, "#") {
			lastTag, err = p.parseTag(line)
			if err != nil {
				return nil, err
			}
			
			// Process special tags
			if err := p.processTag(lastTag); err != nil {
				return nil, err
			}
		} else {
			// Not a tag, so it must be a URI line
			if lastTag != nil && lastTag.Name == TagStreamInf {
				// This is a variant stream URI in a master playlist
				if err := p.processVariantURI(lastTag, line); err != nil {
					return nil, err
				}
				lastTag = nil
			} else {
				// This is a segment URI in a media playlist
				if err := p.processSegmentURI(lastTag, line); err != nil {
					return nil, err
				}
				lastTag = nil
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	// If we have at least one variant, it's a master playlist
	// If we have at least one segment, it's a media playlist
	if len(p.playlist.Master.Variants) > 0 {
		p.playlist.Type = PlaylistTypeMaster
	} else if len(p.playlist.Media.Segments) > 0 {
		p.playlist.Type = PlaylistTypeMedia
	}
	
	return p.playlist, nil
}

// parseTag parses an HLS tag into a Tag structure
func (p *Parser) parseTag(line string) (*Tag, error) {
	tag := &Tag{
		RawLine: line,
	}
	
	// Check if tag has a value
	colonIndex := strings.Index(line, ":")
	if colonIndex == -1 {
		// Simple tag without value
		tag.Name = line
		return tag, nil
	}
	
	// Split tag name and value
	tag.Name = line[:colonIndex]
	tag.Value = line[colonIndex+1:]
	
	// For tags with attributes, parse them
	if tag.Name == TagStreamInf || tag.Name == TagMedia || 
	   tag.Name == TagIFrameStreamInf || tag.Name == TagKey ||
	   tag.Name == TagMap || tag.Name == TagSessionData {
		
		attrs, err := parseAttributes(tag.Value)
		if err != nil {
			return nil, err
		}
		tag.Attributes = attrs
	}
	
	return tag, nil
}

// processTag processes a tag and updates the playlist
func (p *Parser) processTag(tag *Tag) error {
	switch tag.Name {
	case TagVersion:
		// Parse version
		ver, err := strconv.Atoi(tag.Value)
		if err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
		p.playlist.Version = ver
		
	case TagTargetDuration:
		// Parse target duration
		dur, err := strconv.ParseFloat(tag.Value, 64)
		if err != nil {
			return fmt.Errorf("invalid target duration: %w", err)
		}
		p.playlist.Media.TargetDuration = dur
		p.playlist.Type = PlaylistTypeMedia
		
	case TagMediaSequence:
		// Parse media sequence
		seq, err := strconv.ParseUint(tag.Value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid media sequence: %w", err)
		}
		p.playlist.Media.MediaSequence = seq
		p.playlist.Type = PlaylistTypeMedia
		
	case TagDiscontinuitySequence:
		// Parse discontinuity sequence
		seq, err := strconv.ParseUint(tag.Value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid discontinuity sequence: %w", err)
		}
		p.playlist.Media.DiscontinuitySeq = seq
		p.playlist.Type = PlaylistTypeMedia
		
	case TagEndList:
		// Mark playlist as ended
		p.playlist.Media.EndList = true
		p.playlist.Type = PlaylistTypeMedia
		
	case TagAllowCache:
		// Parse allow cache
		p.playlist.Media.AllowCache = tag.Value != "NO"
		p.playlist.Type = PlaylistTypeMedia
		
	case TagPlaylistType:
		// Set playlist type
		p.playlist.Media.PlaylistType = tag.Value
		p.playlist.Type = PlaylistTypeMedia
		
	case TagIFramesOnly:
		// Mark playlist as I-frames only
		p.playlist.Media.IFramesOnly = true
		p.playlist.Type = PlaylistTypeMedia
		
	case TagIndependentSegments:
		// Mark playlist as having independent segments
		if p.playlist.Type == PlaylistTypeMaster || p.playlist.Type == PlaylistTypeUnknown {
			p.playlist.Master.HasIndependentSegments = true
		} else {
			p.playlist.Media.HasIndependentSegments = true
		}
		
	case TagMedia:
		// Add media group
		if err := p.processMediaGroup(tag); err != nil {
			return err
		}
		p.playlist.Type = PlaylistTypeMaster
		
	case TagIFrameStreamInf:
		// Add I-frame stream
		if err := p.processIFrameStream(tag); err != nil {
			return err
		}
		p.playlist.Type = PlaylistTypeMaster
		
	case TagSessionData:
		// Add session data
		if err := p.processSessionData(tag); err != nil {
			return err
		}
		p.playlist.Type = PlaylistTypeMaster
		
	case TagStreamInf:
		// Tag will be processed with the URI line
		p.playlist.Type = PlaylistTypeMaster
		
	case TagInf:
		// Will be processed with the URI line
		p.playlist.Type = PlaylistTypeMedia
		
	case TagDiscontinuity, TagKey, TagByteRange, TagProgramDateTime, TagMap:
		// These will be processed with the next segment
		p.playlist.Type = PlaylistTypeMedia
	}
	
	// Store the tag
	p.playlist.Tags = append(p.playlist.Tags, *tag)
	
	return nil
}

// processVariantURI processes a variant URI line in a master playlist
func (p *Parser) processVariantURI(tag *Tag, uri string) error {
	if tag.Name != TagStreamInf {
		return fmt.Errorf("expected EXT-X-STREAM-INF tag before URI, got %s", tag.Name)
	}
	
	// Get bandwidth
	bandwidth, err := parseAttributeUint(tag.Attributes, AttrBandwidth)
	if err != nil {
		return err
	}
	
	// Add variant
	p.playlist.AddVariant(uri, bandwidth, tag.Attributes)
	
	return nil
}

// processSegmentURI processes a segment URI line in a media playlist
func (p *Parser) processSegmentURI(tag *Tag, uri string) error {
	// If this URI doesn't follow an EXTINF tag, it's invalid
	if tag == nil || tag.Name != TagInf {
		return fmt.Errorf("segment URI must follow EXTINF tag")
	}
	
	// Parse duration and title
	duration, title, err := parseInfValue(tag.Value)
	if err != nil {
		return err
	}
	
	// Add segment
	p.playlist.AddSegment(uri, duration, title)
	
	return nil
}

// processMediaGroup processes a media group tag
func (p *Parser) processMediaGroup(tag *Tag) error {
	typeVal, ok := tag.Attributes[AttrType]
	if !ok {
		return fmt.Errorf("missing TYPE attribute in EXT-X-MEDIA")
	}
	
	groupID, ok := tag.Attributes[AttrGroupID]
	if !ok {
		return fmt.Errorf("missing GROUP-ID attribute in EXT-X-MEDIA")
	}
	
	// Create media group
	group := MediaGroup{
		Type:          typeVal,
		GroupID:       groupID,
		RawAttributes: tag.Value,
	}
	
	// Set optional attributes
	if name, ok := tag.Attributes[AttrName]; ok {
		group.Name = name
	}
	
	if uri, ok := tag.Attributes[AttrURI]; ok {
		group.URI = uri
	}
	
	if lang, ok := tag.Attributes[AttrLanguage]; ok {
		group.Language = lang
	}
	
	if assocLang, ok := tag.Attributes[AttrAssocLanguage]; ok {
		group.AssocLanguage = assocLang
	}
	
	if dflt, ok := tag.Attributes[AttrDefault]; ok {
		group.Default = dflt == "YES"
	}
	
	if auto, ok := tag.Attributes[AttrAutoselect]; ok {
		group.Autoselect = auto == "YES"
	}
	
	if forced, ok := tag.Attributes[AttrForced]; ok {
		group.Forced = forced == "YES"
	}
	
	if instream, ok := tag.Attributes[AttrInstreamID]; ok {
		group.InstreamID = instream
	}
	
	if chars, ok := tag.Attributes[AttrCharacteristics]; ok {
		group.Characteristics = chars
	}
	
	if channels, ok := tag.Attributes[AttrChannels]; ok {
		group.Channels = channels
	}
	
	// Add to the appropriate group type
	if _, ok := p.playlist.Master.MediaGroups[typeVal]; !ok {
		p.playlist.Master.MediaGroups[typeVal] = make([]MediaGroup, 0)
	}
	p.playlist.Master.MediaGroups[typeVal] = append(p.playlist.Master.MediaGroups[typeVal], group)
	
	return nil
}

// processIFrameStream processes an I-frame stream tag
func (p *Parser) processIFrameStream(tag *Tag) error {
	uri, ok := tag.Attributes[AttrURI]
	if !ok {
		return fmt.Errorf("missing URI attribute in EXT-X-I-FRAME-STREAM-INF")
	}
	
	bandwidth, err := parseAttributeUint(tag.Attributes, AttrBandwidth)
	if err != nil {
		return err
	}
	
	// Create I-frame stream
	iframe := IFrameStream{
		URI:           uri,
		Bandwidth:     bandwidth,
		RawAttributes: tag.Value,
	}
	
	// Set optional attributes
	if avgBw, ok := tag.Attributes[AttrAverageBandwidth]; ok {
		if val, err := strconv.ParseUint(avgBw, 10, 64); err == nil {
			iframe.AverageBandwidth = val
		}
	}
	
	if codecs, ok := tag.Attributes[AttrCodecs]; ok {
		iframe.Codecs = codecs
	}
	
	if res, ok := tag.Attributes[AttrResolution]; ok {
		iframe.Resolution = res
	}
	
	if hdcp, ok := tag.Attributes[AttrHDCPLevel]; ok {
		iframe.HDCPLevel = hdcp
	}
	
	if video, ok := tag.Attributes[AttrVideo]; ok {
		iframe.VideoGroup = video
	}
	
	// Add to playlist
	p.playlist.Master.IFrameStreams = append(p.playlist.Master.IFrameStreams, iframe)
	
	return nil
}

// processSessionData processes a session data tag
func (p *Parser) processSessionData(tag *Tag) error {
	dataID, ok := tag.Attributes[AttrDataID]
	if !ok {
		return fmt.Errorf("missing DATA-ID attribute in EXT-X-SESSION-DATA")
	}
	
	// Create session data
	sessData := SessionData{
		DataID:        dataID,
		RawAttributes: tag.Value,
	}
	
	// Set optional attributes
	if value, ok := tag.Attributes[AttrValue]; ok {
		sessData.Value = value
	}
	
	if uri, ok := tag.Attributes[AttrURI]; ok {
		sessData.URI = uri
	}
	
	if lang, ok := tag.Attributes[AttrLanguage]; ok {
		sessData.Language = lang
	}
	
	// Add to playlist
	p.playlist.Master.SessionData = append(p.playlist.Master.SessionData, sessData)
	
	return nil
}

// parseAttributes parses a string of comma-separated attributes
func parseAttributes(s string) (map[string]string, error) {
	attrs := make(map[string]string)
	r := regexp.MustCompile(`([A-Z-]+)=("[^"]*"|[^",]+)`)
	
	matches := r.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}
		
		key := match[1]
		value := match[2]
		
		// Remove quotes if present
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = value[1 : len(value)-1]
		}
		
		attrs[key] = value
	}
	
	return attrs, nil
}

// parseAttributeUint parses a uint64 attribute
func parseAttributeUint(attrs map[string]string, name string) (uint64, error) {
	valStr, ok := attrs[name]
	if !ok {
		return 0, fmt.Errorf("missing %s attribute", name)
	}
	
	val, err := strconv.ParseUint(valStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value: %w", name, err)
	}
	
	return val, nil
}

// parseInfValue parses the value of an EXTINF tag
func parseInfValue(s string) (float64, string, error) {
	parts := strings.SplitN(s, ",", 2)
	
	// Parse duration
	duration, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid EXTINF duration: %w", err)
	}
	
	// Get title if present
	var title string
	if len(parts) > 1 {
		title = parts[1]
	}
	
	return duration, title, nil
}