// Server configuration options
//
// Defines the available server options and their defaults:
// - Listen addresses and ports
// - TLS configuration
// - Timeouts (read, write, idle)
// - Connection limits
// - Keep-alive settings

package server

import (
	"crypto/tls"
	"time"

	"github.com/ilijajolevski/ilinden/internal/config"
)

// Options represents all the HTTP server configuration options
type Options struct {
	Address           string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxHeaderBytes    int
	EnableCompression bool
	TLSConfig         *tls.Config
}

// NewOptionsFromConfig creates server options from configuration
func NewOptionsFromConfig(cfg *config.Config) Options {
	return Options{
		Address:           cfg.GetAddress(),
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ShutdownTimeout:   cfg.Server.ShutdownTimeout,
		MaxHeaderBytes:    cfg.Server.MaxHeaderBytes,
		EnableCompression: cfg.Server.EnableCompression,
		TLSConfig:         nil, // TLS config would be set separately if needed
	}
}

// WithTLS adds TLS configuration to the server options
func (o Options) WithTLS(cert, key string) (Options, error) {
	cert2, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return o, err
	}

	o.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert2},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	return o, nil
}