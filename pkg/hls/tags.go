// HLS tag definitions and constants
package hls

// HLS tag constants
const (
	// HLS version tags
	TagExtM3U       = "#EXTM3U"
	TagVersion      = "#EXT-X-VERSION"
	
	// Master playlist tags
	TagStreamInf        = "#EXT-X-STREAM-INF"
	TagMediaSequence    = "#EXT-X-MEDIA-SEQUENCE"
	TagMedia            = "#EXT-X-MEDIA"
	TagIFrameStreamInf  = "#EXT-X-I-FRAME-STREAM-INF"
	TagSessionData      = "#EXT-X-SESSION-DATA"
	TagIndependentSegments = "#EXT-X-INDEPENDENT-SEGMENTS"
	
	// Media playlist tags
	TagTargetDuration   = "#EXT-X-TARGETDURATION"
	TagInf              = "#EXTINF"
	TagByteRange        = "#EXT-X-BYTERANGE"
	TagDiscontinuity    = "#EXT-X-DISCONTINUITY"
	TagKey              = "#EXT-X-KEY"
	TagMap              = "#EXT-X-MAP"
	TagProgramDateTime  = "#EXT-X-PROGRAM-DATE-TIME"
	TagEndList          = "#EXT-X-ENDLIST"
	TagDiscontinuitySequence = "#EXT-X-DISCONTINUITY-SEQUENCE"
	TagAllowCache       = "#EXT-X-ALLOW-CACHE"
	TagPlaylistType     = "#EXT-X-PLAYLIST-TYPE"
	TagIFramesOnly      = "#EXT-X-I-FRAMES-ONLY"
	
	// Common stream information attributes
	AttrBandwidth       = "BANDWIDTH"
	AttrAverageBandwidth = "AVERAGE-BANDWIDTH"
	AttrCodecs          = "CODECS"
	AttrResolution      = "RESOLUTION"
	AttrFrameRate       = "FRAME-RATE"
	AttrHDCPLevel       = "HDCP-LEVEL"
	AttrAudio           = "AUDIO"
	AttrVideo           = "VIDEO"
	AttrSubtitles       = "SUBTITLES"
	AttrClosedCaptions  = "CLOSED-CAPTIONS"
	AttrURI             = "URI"
	
	// Key attributes
	AttrMethod          = "METHOD"
	AttrKeyFormat       = "KEYFORMAT"
	AttrKeyFormatVersions = "KEYFORMATVERSIONS"
	AttrIV              = "IV"
	
	// Media attributes
	AttrType            = "TYPE"
	AttrGroupID         = "GROUP-ID"
	AttrLanguage        = "LANGUAGE"
	AttrAssocLanguage   = "ASSOC-LANGUAGE"
	AttrName            = "NAME"
	AttrDefault         = "DEFAULT"
	AttrAutoselect      = "AUTOSELECT"
	AttrForced          = "FORCED"
	AttrInstreamID      = "INSTREAM-ID"
	AttrCharacteristics = "CHARACTERISTICS"
	AttrChannels        = "CHANNELS"
	
	// Session data attributes
	AttrDataID          = "DATA-ID"
	AttrValue           = "VALUE"
	AttrLanguage        = "LANGUAGE"
)

// PlaylistType represents the type of playlist (master or media)
type PlaylistType int

const (
	PlaylistTypeUnknown PlaylistType = iota
	PlaylistTypeMaster
	PlaylistTypeMedia
)

// KeyMethod represents the encryption method
type KeyMethod string

const (
	KeyMethodNone    KeyMethod = "NONE"
	KeyMethodAES128  KeyMethod = "AES-128"
	KeyMethodSampleAES KeyMethod = "SAMPLE-AES"
)

// MediaType represents the type of media
type MediaType string

const (
	MediaTypeAudio   MediaType = "AUDIO"
	MediaTypeVideo   MediaType = "VIDEO"
	MediaTypeSubtitles MediaType = "SUBTITLES"
	MediaTypeClosedCaptions MediaType = "CLOSED-CAPTIONS"
)