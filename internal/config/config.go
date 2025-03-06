// Configuration structure definitions
//
// Defines all configuration options as structured Go types
// with validation tags and defaults
//
// Main sections:
// - ServerConfig: HTTP server settings
// - OriginConfig: Origin server connection settings
// - JWTConfig: JWT validation parameters
// - CacheConfig: Caching behavior settings
// - RedisConfig: Optional Redis connection
// - LogConfig: Logging parameters
// - MetricsConfig: Telemetry settings

package config

import (
	"time"
)

// Config represents the top-level configuration structure
type Config struct {
	Server   ServerConfig   `yaml:"server" json:"server"`
	Origin   OriginConfig   `yaml:"origin" json:"origin"`
	JWT      JWTConfig      `yaml:"jwt" json:"jwt"`
	Cache    CacheConfig    `yaml:"cache" json:"cache"`
	Redis    RedisConfig    `yaml:"redis" json:"redis"`
	Log      LogConfig      `yaml:"log" json:"log"`
	Metrics  MetricsConfig  `yaml:"metrics" json:"metrics"`
	Tracing  TracingConfig  `yaml:"tracing" json:"tracing"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Host              string        `yaml:"host" json:"host" default:"0.0.0.0"`
	Port              int           `yaml:"port" json:"port" default:"8080"`
	ReadTimeout       time.Duration `yaml:"readTimeout" json:"readTimeout" default:"5s"`
	WriteTimeout      time.Duration `yaml:"writeTimeout" json:"writeTimeout" default:"10s"`
	IdleTimeout       time.Duration `yaml:"idleTimeout" json:"idleTimeout" default:"120s"`
	ShutdownTimeout   time.Duration `yaml:"shutdownTimeout" json:"shutdownTimeout" default:"30s"`
	MaxHeaderBytes    int           `yaml:"maxHeaderBytes" json:"maxHeaderBytes" default:"1048576"` // 1MB
	MaxRequestBodyMB  int           `yaml:"maxRequestBodyMB" json:"maxRequestBodyMB" default:"10"`
	EnableCompression bool          `yaml:"enableCompression" json:"enableCompression" default:"true"`
	TrustedProxies    []string      `yaml:"trustedProxies" json:"trustedProxies"`
}

// OriginConfig contains settings for communicating with origin servers
type OriginConfig struct {
	Timeout               time.Duration `yaml:"timeout" json:"timeout" default:"5s"`
	MaxIdleConns          int           `yaml:"maxIdleConns" json:"maxIdleConns" default:"100"`
	MaxIdleConnsPerHost   int           `yaml:"maxIdleConnsPerHost" json:"maxIdleConnsPerHost" default:"10"`
	MaxConnsPerHost       int           `yaml:"maxConnsPerHost" json:"maxConnsPerHost" default:"100"`
	IdleConnTimeout       time.Duration `yaml:"idleConnTimeout" json:"idleConnTimeout" default:"90s"`
	TLSHandshakeTimeout   time.Duration `yaml:"tlsHandshakeTimeout" json:"tlsHandshakeTimeout" default:"10s"`
	ExpectContinueTimeout time.Duration `yaml:"expectContinueTimeout" json:"expectContinueTimeout" default:"1s"`
	DefaultScheme         string        `yaml:"defaultScheme" json:"defaultScheme" default:"https"`
	BaseURL               string        `yaml:"baseURL" json:"baseURL"`
	RetryCount            int           `yaml:"retryCount" json:"retryCount" default:"3"`
	RetryWaitMin          time.Duration `yaml:"retryWaitMin" json:"retryWaitMin" default:"100ms"`
	RetryWaitMax          time.Duration `yaml:"retryWaitMax" json:"retryWaitMax" default:"2s"`
	CircuitBreaker        bool          `yaml:"circuitBreaker" json:"circuitBreaker" default:"true"`
}

// JWTConfig contains JWT validation parameters
type JWTConfig struct {
	Enabled         bool     `yaml:"enabled" json:"enabled" default:"true"`
	ParamName       string   `yaml:"paramName" json:"paramName" default:"token"`
	HeaderName      string   `yaml:"headerName" json:"headerName" default:"Authorization"`
	Secret          string   `yaml:"secret" json:"secret"`
	KeysURL         string   `yaml:"keysUrl" json:"keysUrl"`
	RequiredClaims  []string `yaml:"requiredClaims" json:"requiredClaims"`
	ClaimsNamespace string   `yaml:"claimsNamespace" json:"claimsNamespace"`
	Issuer          string   `yaml:"issuer" json:"issuer"`
	Audience        string   `yaml:"audience" json:"audience"`
	AllowedAlgs     []string `yaml:"allowedAlgs" json:"allowedAlgs" default:"[\"HS256\", \"RS256\"]"`
}

// CacheConfig contains caching behavior settings
type CacheConfig struct {
	Enabled            bool          `yaml:"enabled" json:"enabled" default:"true"`
	TTLMaster          time.Duration `yaml:"ttlMaster" json:"ttlMaster" default:"10s"`
	TTLMedia           time.Duration `yaml:"ttlMedia" json:"ttlMedia" default:"2s"`
	MaxSize            int           `yaml:"maxSize" json:"maxSize" default:"10000"`
	ShardCount         int           `yaml:"shardCount" json:"shardCount" default:"16"`
	StaleWhileRevalidate bool         `yaml:"staleWhileRevalidate" json:"staleWhileRevalidate" default:"true"`
	UseRedis           bool          `yaml:"useRedis" json:"useRedis" default:"false"`
}

// RedisConfig contains optional Redis connection details
type RedisConfig struct {
	Enabled        bool          `yaml:"enabled" json:"enabled" default:"false"`
	Addresses      []string      `yaml:"addresses" json:"addresses" default:"[\"localhost:6379\"]"`
	Password       string        `yaml:"password" json:"password"`
	DB             int           `yaml:"db" json:"db" default:"0"`
	PoolSize       int           `yaml:"poolSize" json:"poolSize" default:"10"`
	MinIdleConns   int           `yaml:"minIdleConns" json:"minIdleConns" default:"5"`
	DialTimeout    time.Duration `yaml:"dialTimeout" json:"dialTimeout" default:"5s"`
	ReadTimeout    time.Duration `yaml:"readTimeout" json:"readTimeout" default:"3s"`
	WriteTimeout   time.Duration `yaml:"writeTimeout" json:"writeTimeout" default:"3s"`
	PoolTimeout    time.Duration `yaml:"poolTimeout" json:"poolTimeout" default:"4s"`
	IdleTimeout    time.Duration `yaml:"idleTimeout" json:"idleTimeout" default:"5m"`
	MaxConnAge     time.Duration `yaml:"maxConnAge" json:"maxConnAge" default:"30m"`
	TrackingPrefix string        `yaml:"trackingPrefix" json:"trackingPrefix" default:"ilinden:player:"`
	TrackingExpiry time.Duration `yaml:"trackingExpiry" json:"trackingExpiry" default:"5m"`
}

// LogConfig contains logging parameters
type LogConfig struct {
	Level       string `yaml:"level" json:"level" default:"info"`
	Format      string `yaml:"format" json:"format" default:"json"`
	OutputPath  string `yaml:"outputPath" json:"outputPath" default:"stdout"`
	ErrorPath   string `yaml:"errorPath" json:"errorPath" default:"stderr"`
	Development bool   `yaml:"development" json:"development" default:"false"`
}

// MetricsConfig contains telemetry settings
type MetricsConfig struct {
	Enabled       bool   `yaml:"enabled" json:"enabled" default:"true"`
	Address       string `yaml:"address" json:"address" default:":9090"`
	Path          string `yaml:"path" json:"path" default:"/metrics"`
	CollectSystem bool   `yaml:"collectSystem" json:"collectSystem" default:"true"`
}

// TracingConfig contains distributed tracing settings
type TracingConfig struct {
	Enabled     bool    `yaml:"enabled" json:"enabled" default:"false"`
	ServiceName string  `yaml:"serviceName" json:"serviceName" default:"ilinden"`
	Endpoint    string  `yaml:"endpoint" json:"endpoint" default:"localhost:4317"`
	SampleRate  float64 `yaml:"sampleRate" json:"sampleRate" default:"0.1"`
}