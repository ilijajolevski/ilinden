// Main entry point for the Ilinden HLS proxy server
//
// Responsibilities:
// - Parse command line flags
// - Load and validate configuration
// - Set up signal handling for graceful shutdown
// - Initialize logging and metrics
// - Start the server
// - Wait for termination signals

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ilijajolevski/ilinden/internal/api"
	"github.com/ilijajolevski/ilinden/internal/cache"
	"github.com/ilijajolevski/ilinden/internal/config"
	"github.com/ilijajolevski/ilinden/internal/middleware"
	"github.com/ilijajolevski/ilinden/internal/proxy"
	"github.com/ilijajolevski/ilinden/internal/redis"
	"github.com/ilijajolevski/ilinden/internal/server"
	"github.com/ilijajolevski/ilinden/internal/telemetry"
)

var (
	// Command line flags
	configPath    = flag.String("config", "", "Path to configuration file")
	configDefault = flag.String("config-default", "configs/ilinden.yaml", "Path to default configuration file")
	version       = flag.Bool("version", false, "Print version and exit")
)

// Version information
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Parse command line flags
	flag.Parse()

	// Print version and exit if requested
	if *version {
		fmt.Printf("Ilinden HLS Proxy v%s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		os.Exit(0)
	}

	// Determine configuration path
	cfgPath := *configPath
	if cfgPath == "" {
		cfgPath = *configDefault
	}

	// Make sure path is absolute
	if !filepath.IsAbs(cfgPath) {
		absPath, err := filepath.Abs(cfgPath)
		if err != nil {
			log.Fatalf("Failed to get absolute path: %v", err)
		}
		cfgPath = absPath
	}

	// Load configuration
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize logging
	logger := telemetry.NewLogger(cfg.Log.Level, cfg.Log.Format, cfg.Log.OutputPath)
	logger.Info("Starting Ilinden HLS Proxy", "version", Version, "commit", GitCommit)

	// Initialize metrics
	metrics := telemetry.NewMetrics()

	// Initialize cache
	var cacheImpl cache.Cache
	if cfg.Cache.Enabled {
		cacheOpts := cache.MemoryOptions{
			MaxSize:   cfg.Cache.MaxSize,
			ShardSize: cfg.Cache.ShardCount,
		}
		cacheImpl = cache.NewMemoryWithOptions(cacheOpts)
		logger.Info("Initialized memory cache", "maxSize", cfg.Cache.MaxSize, "shards", cfg.Cache.ShardCount)
	} else {
		logger.Info("Cache disabled")
	}

	// Initialize Redis client if enabled
	var redisTracker *redis.Tracker
	if cfg.Redis.Enabled {
		// In a real implementation, this would initialize the Redis client
		logger.Info("Redis tracking enabled")
	} else {
		logger.Info("Redis tracking disabled")
	}

	// Create router
	mux := http.NewServeMux()

	// Create proxy handler
	proxyHandler := proxy.NewHandler(proxy.HandlerOptions{
		Config:       cfg,
		Cache:        cacheImpl,
		Logger:       logger,
		Metrics:      metrics,
		RedisTracker: redisTracker,
	})

	// Setup middleware chain
	chain := middleware.NewChain(
		middleware.Recovery(logger),
		middleware.Logging(logger),
		middleware.Metrics(metrics),
	)

	// Register routes
	mux.Handle("/", chain.Then(proxyHandler))

	// Register health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		api.WriteResponse(w, http.StatusOK, api.NewResponse(true, "OK", nil))
	})

	// Register metrics endpoint if enabled
	if cfg.Metrics.Enabled {
		mux.HandleFunc(cfg.Metrics.Path, func(w http.ResponseWriter, r *http.Request) {
			// This would typically expose Prometheus metrics
			// For our simple implementation, we'll just return some basic stats
			if m, ok := metrics.(*telemetry.SimpleMetrics); ok {
				api.WriteJSON(w, http.StatusOK, m.DumpMetrics())
			} else {
				api.WriteResponse(w, http.StatusOK, api.NewResponse(true, "Metrics not available", nil))
			}
		})
	}

	// Create and configure the server
	srv := server.New(
		server.NewOptionsFromConfig(cfg),
		mux,
	)

	// Setup graceful shutdown
	shutdown := server.NewGracefulShutdown(srv, cfg.Server.ShutdownTimeout)

	// Start the server
	logger.Info("Starting server", "address", cfg.GetAddress())
	if err := srv.Start(); err != nil {
		logger.Error("Failed to start server", "error", err.Error())
		os.Exit(1)
	}

	// Wait for shutdown signal
	shutdown.WaitForShutdown()

	// Perform any cleanup
	if cacheImpl != nil {
		logger.Info("Cleaning up cache")
		cacheImpl.Clear()
	}

	logger.Info("Server shutdown complete")
}
