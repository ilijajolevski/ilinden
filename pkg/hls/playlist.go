// HLS playlist data structures
//
// Type definitions for HLS playlists:
// - Master playlist structure
// - Media playlist structure
// - Segment information
// - Tag representation

package hls

import (
	"fmt"
	"strconv"
	"strings"
)

// Playlist represents an HLS playlist (either master or media)
type Playlist struct {
	Type           PlaylistType
	Version        int
	Tags           []Tag
	Master         MasterPlaylist
	Media          MediaPlaylist
	OriginalHeader string
	RawLines       []string
}

// MasterPlaylist contains data specific to master playlists
type MasterPlaylist struct {
	Variants       []Variant
	MediaGroups    map[string][]MediaGroup
	IFrameStreams  []IFrameStream
	SessionData    []SessionData
	HasIndependentSegments bool
}

// MediaPlaylist contains data specific to media playlists
type MediaPlaylist struct {
	TargetDuration     float64
	MediaSequence      uint64
	Segments           []Segment
	EndList            bool
	DiscontinuitySeq   uint64
	AllowCache         bool
	PlaylistType       string
	IFramesOnly        bool
	HasIndependentSegments bool
}

// Variant represents a stream variant in a master playlist
type Variant struct {
	URI                 string
	Bandwidth           uint64
	AverageBandwidth    uint64
	Codecs              string
	Resolution          string
	FrameRate           float64
	HDCPLevel           string
	AudioGroup          string
	VideoGroup          string
	SubtitlesGroup      string
	ClosedCaptionsGroup string
	RawAttributes       string
}

// MediaGroup represents a media group in a master playlist
type MediaGroup struct {
	Type            string
	GroupID         string
	Name            string
	URI             string
	Language        string
	AssocLanguage   string
	Default         bool
	Autoselect      bool
	Forced          bool
	InstreamID      string
	Characteristics string
	Channels        string
	RawAttributes   string
}

// IFrameStream represents an I-frame stream in a master playlist
type IFrameStream struct {
	URI                 string
	Bandwidth           uint64
	AverageBandwidth    uint64
	Codecs              string
	Resolution          string
	HDCPLevel           string
	VideoGroup          string
	RawAttributes       string
}

// SessionData represents session data in a master playlist
type SessionData struct {
	DataID          string
	Value           string
	URI             string
	Language        string
	RawAttributes   string
}

// Segment represents a media segment in a media playlist
type Segment struct {
	URI                string
	Duration           float64
	Title              string
	ByteRange          string
	Discontinuity      bool
	ProgramDateTime    string
	Key                *Key
	Map                *Map
}

// Key represents an encryption key for segments
type Key struct {
	Method           KeyMethod
	URI              string
	IV               string
	KeyFormat        string
	KeyFormatVersions string
	RawAttributes    string
}

// Map represents a segment map
type Map struct {
	URI              string
	ByteRange        string
	RawAttributes    string
}

// Tag represents a parsed HLS tag with its attributes
type Tag struct {
	Name         string
	Value        string
	Attributes   map[string]string
	RawLine      string
}

// NewPlaylist creates a new HLS playlist
func NewPlaylist() *Playlist {
	return &Playlist{
		Type:    PlaylistTypeUnknown,
		Version: 1, // Default version
		Tags:    make([]Tag, 0),
		Master: MasterPlaylist{
			Variants:    make([]Variant, 0),
			MediaGroups: make(map[string][]MediaGroup),
			IFrameStreams: make([]IFrameStream, 0),
			SessionData: make([]SessionData, 0),
		},
		Media: MediaPlaylist{
			Segments: make([]Segment, 0),
		},
		RawLines: make([]string, 0),
	}
}

// String returns the playlist as a string
func (p *Playlist) String() string {
	var sb strings.Builder
	
	// Write header
	sb.WriteString(TagExtM3U + "\n")
	sb.WriteString(fmt.Sprintf("%s:%d\n", TagVersion, p.Version))
	
	// Write other global tags
	for _, tag := range p.Tags {
		if tag.Name != TagExtM3U && tag.Name != TagVersion {
			sb.WriteString(tag.String() + "\n")
		}
	}
	
	// Write playlist-specific content
	if p.Type == PlaylistTypeMaster {
		// Write master playlist
		
		// Independent segments if present
		if p.Master.HasIndependentSegments {
			sb.WriteString(TagIndependentSegments + "\n")
		}
		
		// Media groups
		for _, groups := range p.Master.MediaGroups {
			for _, group := range groups {
				sb.WriteString(fmt.Sprintf("%s:%s\n", TagMedia, group.RawAttributes))
			}
		}
		
		// Session data
		for _, data := range p.Master.SessionData {
			sb.WriteString(fmt.Sprintf("%s:%s\n", TagSessionData, data.RawAttributes))
		}
		
		// Variants
		for _, variant := range p.Master.Variants {
			sb.WriteString(fmt.Sprintf("%s:%s\n%s\n", TagStreamInf, variant.RawAttributes, variant.URI))
		}
		
		// I-frame streams
		for _, iframe := range p.Master.IFrameStreams {
			sb.WriteString(fmt.Sprintf("%s:%s\n", TagIFrameStreamInf, iframe.RawAttributes))
		}
		
	} else if p.Type == PlaylistTypeMedia {
		// Write media playlist
		
		// Independent segments if present
		if p.Media.HasIndependentSegments {
			sb.WriteString(TagIndependentSegments + "\n")
		}
		
		// Target duration
		sb.WriteString(fmt.Sprintf("%s:%d\n", TagTargetDuration, int(p.Media.TargetDuration)))
		
		// Media sequence
		sb.WriteString(fmt.Sprintf("%s:%d\n", TagMediaSequence, p.Media.MediaSequence))
		
		// Discontinuity sequence if non-zero
		if p.Media.DiscontinuitySeq > 0 {
			sb.WriteString(fmt.Sprintf("%s:%d\n", TagDiscontinuitySequence, p.Media.DiscontinuitySeq))
		}
		
		// Allow cache if specified
		if !p.Media.AllowCache {
			sb.WriteString(fmt.Sprintf("%s:NO\n", TagAllowCache))
		}
		
		// Playlist type if specified
		if p.Media.PlaylistType != "" {
			sb.WriteString(fmt.Sprintf("%s:%s\n", TagPlaylistType, p.Media.PlaylistType))
		}
		
		// I-frames only if specified
		if p.Media.IFramesOnly {
			sb.WriteString(fmt.Sprintf("%s\n", TagIFramesOnly))
		}
		
		// Segments
		for _, segment := range p.Media.Segments {
			// Key information if present
			if segment.Key != nil {
				sb.WriteString(fmt.Sprintf("%s:%s\n", TagKey, segment.Key.RawAttributes))
			}
			
			// Map information if present
			if segment.Map != nil {
				sb.WriteString(fmt.Sprintf("%s:%s\n", TagMap, segment.Map.RawAttributes))
			}
			
			// Program date time if present
			if segment.ProgramDateTime != "" {
				sb.WriteString(fmt.Sprintf("%s:%s\n", TagProgramDateTime, segment.ProgramDateTime))
			}
			
			// Discontinuity if present
			if segment.Discontinuity {
				sb.WriteString(fmt.Sprintf("%s\n", TagDiscontinuity))
			}
			
			// Byte range if present
			if segment.ByteRange != "" {
				sb.WriteString(fmt.Sprintf("%s:%s\n", TagByteRange, segment.ByteRange))
			}
			
			// Segment information
			if segment.Title != "" {
				sb.WriteString(fmt.Sprintf("%s:%.3f,%s\n", TagInf, segment.Duration, segment.Title))
			} else {
				sb.WriteString(fmt.Sprintf("%s:%.3f\n", TagInf, segment.Duration))
			}
			
			// URI
			sb.WriteString(segment.URI + "\n")
		}
		
		// End list if specified
		if p.Media.EndList {
			sb.WriteString(fmt.Sprintf("%s\n", TagEndList))
		}
	}
	
	return sb.String()
}

// String returns a tag as a string
func (t *Tag) String() string {
	if t.Value != "" {
		return fmt.Sprintf("%s:%s", t.Name, t.Value)
	}
	return t.Name
}

// IsMaster returns true if the playlist is a master playlist
func (p *Playlist) IsMaster() bool {
	return p.Type == PlaylistTypeMaster
}

// IsMedia returns true if the playlist is a media playlist
func (p *Playlist) IsMedia() bool {
	return p.Type == PlaylistTypeMedia
}

// AddVariant adds a variant to a master playlist
func (p *Playlist) AddVariant(uri string, bandwidth uint64, attrs map[string]string) {
	v := Variant{
		URI:       uri,
		Bandwidth: bandwidth,
	}
	
	// Set other attributes if provided
	if avgBw, ok := attrs[AttrAverageBandwidth]; ok {
		if val, err := strconv.ParseUint(avgBw, 10, 64); err == nil {
			v.AverageBandwidth = val
		}
	}
	
	if codecs, ok := attrs[AttrCodecs]; ok {
		v.Codecs = codecs
	}
	
	if res, ok := attrs[AttrResolution]; ok {
		v.Resolution = res
	}
	
	if fr, ok := attrs[AttrFrameRate]; ok {
		if val, err := strconv.ParseFloat(fr, 64); err == nil {
			v.FrameRate = val
		}
	}
	
	if hdcp, ok := attrs[AttrHDCPLevel]; ok {
		v.HDCPLevel = hdcp
	}
	
	if audio, ok := attrs[AttrAudio]; ok {
		v.AudioGroup = audio
	}
	
	if video, ok := attrs[AttrVideo]; ok {
		v.VideoGroup = video
	}
	
	if subs, ok := attrs[AttrSubtitles]; ok {
		v.SubtitlesGroup = subs
	}
	
	if cc, ok := attrs[AttrClosedCaptions]; ok {
		v.ClosedCaptionsGroup = cc
	}
	
	// Build raw attributes string
	var parts []string
	parts = append(parts, fmt.Sprintf("%s=%d", AttrBandwidth, bandwidth))
	
	for k, v := range attrs {
		if k != AttrBandwidth {
			// Quote string values
			if k == AttrCodecs || k == AttrResolution || 
			   k == AttrAudio || k == AttrVideo || 
			   k == AttrSubtitles || k == AttrClosedCaptions ||
			   k == AttrHDCPLevel {
				parts = append(parts, fmt.Sprintf("%s=\"%s\"", k, v))
			} else {
				parts = append(parts, fmt.Sprintf("%s=%s", k, v))
			}
		}
	}
	
	v.RawAttributes = strings.Join(parts, ",")
	
	p.Master.Variants = append(p.Master.Variants, v)
	p.Type = PlaylistTypeMaster
}

// AddSegment adds a segment to a media playlist
func (p *Playlist) AddSegment(uri string, duration float64, title string) {
	s := Segment{
		URI:      uri,
		Duration: duration,
		Title:    title,
	}
	
	p.Media.Segments = append(p.Media.Segments, s)
	p.Type = PlaylistTypeMedia
}

// SetTargetDuration sets the target duration for a media playlist
func (p *Playlist) SetTargetDuration(duration float64) {
	p.Media.TargetDuration = duration
	p.Type = PlaylistTypeMedia
}

// SetEndList marks a media playlist as complete (VOD)
func (p *Playlist) SetEndList() {
	p.Media.EndList = true
	p.Type = PlaylistTypeMedia
}

// SetMediaSequence sets the media sequence number for a media playlist
func (p *Playlist) SetMediaSequence(sequence uint64) {
	p.Media.MediaSequence = sequence
	p.Type = PlaylistTypeMedia
}