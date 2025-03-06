# Ilinden Development Guide

## Build & Run Commands
```bash
go build -o bin/ilinden ./cmd/ilinden
go run ./cmd/ilinden

# Testing
go test ./...                      # Run all tests
go test ./internal/cache/...       # Test specific package
go test -run TestName ./internal/  # Run specific test

# Linting & Code Check
go vet ./...                       # Run static analysis
gofmt -s -w .                      # Format code (or use goimports)
go mod tidy                        # Clean up dependencies
```

## Code Style Guidelines
- **Imports**: Group standard library first, then external, then internal
- **Formatting**: Follow Go standard formatting (gofmt)
- **Error Handling**: Always check errors, provide context with wrapped errors
- **Documentation**: All exported functions and packages must have comments
- **Naming**: 
  - CamelCase (not snake_case)
  - Short, descriptive variable names
  - Acronyms uppercase (HTTP, URL)
- **Package Structure**: Follow standard Go project layout
- **Tests**: Use table-driven tests, aim for >80% coverage

## Architecture
Ilinden is a high-performance HTTP proxy for HLS streaming with JWT token propagation. 
See detailed architecture in `docs/architecture.md`.

Key components:
- HTTP proxy with HLS playlist processing
- JWT token extraction and injection
- Memory-efficient caching system
- Optional Redis-based player tracking
- Prometheus metrics and structured logging

## Core Purpose

The server proxies HTTP calls for HLS (variant) playlists and chunklists from media players to origin servers. It extracts JWT tokens from initial requests and injects them into every URL within the playlists returned to clients. This allows for secure, per-player token propagation without requiring the origin server to manage token continuation.

## Key Design Goals

1. **High Performance**: Optimized for high throughput and low latency
2. **Horizontal Scalability**: Stateless design for easy scaling
3. **Resilience**: Fault-tolerant with connection pooling and circuit breaking
4. **Efficient Caching**: Intelligent caching system for live content
5. **Low Resource Footprint**: Memory-efficient parsing and processing
6. **Optional Player Tracking**: Non-blocking activity monitoring

## System Components

### HTTP Server Engine
- High-performance HTTP server built on Go's `net/http`
- Connection pooling and keep-alive optimization
- Graceful shutdown with request draining
- Configurable timeouts and resource limits

### Request Router & Middleware
- Path-based routing with fast matching
- Middleware chain for cross-cutting concerns
- Specialized handlers for different playlist types
- Optional Redis-based player tracking

### JWT Processing
- Token extraction from URL parameters
- Validation against configurable rules
- Stateless token handling
- Token propagation to all playlist URLs

### Playlist Processing
- Efficient M3U8 parsing with minimal allocations
- URL rewriting for both master and media playlists
- Conversion between relative and absolute URLs
- Preservation of all HLS tags and attributes

### Caching Subsystem
- Multi-level caching with memory-conscious eviction
- Short TTLs for live content
- Stale-while-revalidate pattern
- Separate strategies for master and media playlists

### Origin Communication
- Connection pooling to origin servers
- Exponential backoff for retries
- Circuit breaking for failing origins
- Timeout and error handling

### Player Tracking (Optional)
- Redis-based player activity monitoring
- Asynchronous, non-blocking design
- Aggregated metrics for concurrent viewers
- Clean failure handling if Redis is unavailable

### Observability
- Prometheus metrics
- Structured logging
- Health check endpoints
- Optional distributed tracing

## Request Flow

1. Player sends initial request with JWT in URL parameter
2. Proxy extracts and validates JWT token
3. Request is forwarded to origin server
4. Origin returns playlist content
5. Proxy parses playlist and modifies URLs:
  - For master playlists: URLs point back to proxy with JWT tokens
  - For media playlists: Segment URLs point directly to origin with JWT tokens
6. Modified playlist is returned to player
7. For media segments, player requests content directly from origin with JWT

## Configuration System

- YAML-based configuration with environment variable overrides
- Dynamic reloading capability
- Comprehensive options for all components
- Feature flags for optional capabilities

## Performance Considerations

- Minimal memory allocations
- Buffer pooling for request/response bodies
- Non-blocking operations where possible
- Efficient playlist parsing
- Intelligent caching for high hit rates

## Deployment Architecture

- Containerized deployment
- Horizontal scaling with any number of instances
- No sticky sessions required
- Kubernetes-ready with health checks and metrics

## Redis Integration (for Player Tracking)

- Connection pooling to Redis
- Circuit breaking for Redis failures
- Efficient data structures for player activity
- Time-based expiration of inactive players
- Non-blocking queue for tracking events

## Security Aspects

- JWT signature verification
- Input sanitization and validation
- TLS configuration with modern ciphers
- Rate limiting capabilities
- Origin request protection

## File System Organization

The project follows a clean architecture with separation of concerns:

- `cmd/`: Application entry points
- `configs/`: Configuration templates and examples
- `internal/`: Internal packages not meant for external use
  - `config/`: Configuration parsing and validation
  - `server/`: HTTP server implementation
  - `proxy/`: Core proxy functionality
  - `playlist/`: HLS playlist handling
  - `jwt/`: JWT processing
  - `cache/`: Caching system
  - `middleware/`: HTTP middleware components
  - `redis/`: Redis client and tracking
  - `telemetry/`: Metrics, logging, and tracing
  - `api/`: Management API endpoints
  - `utils/`: Common utilities
- `pkg/`: Public packages that could be used by other applications
  - `hls/`: HLS protocol handling
  - `jwtheader/`: JWT extraction utilities
- `build/`: Build and deployment configurations
- `scripts/`: Utility scripts
- `docs/`: Documentation

## Implementation Guidelines

- Go language (1.21+)
- Minimal external dependencies
- Thorough error handling
- Comprehensive test coverage
- Performance benchmarking
- Clear documentation

## Operational Characteristics

- Graceful startup and shutdown
- Low CPU and memory footprint
- Predictable scaling behavior
- Comprehensive metrics for monitoring
- Clear error logging

## Intended Use Cases

- Live HLS streaming with authentication
- Multi-tenant streaming platforms
- Geographically distributed content delivery
- High-volume streaming services
- Services requiring viewer analytics

## Non-Functional Requirements

- Proxy latency under 5ms for cached content
- Support for thousands of concurrent connections per instance
- 99.99% availability with proper deployment
- Graceful handling of origin server failures
- Minimal impact from player tracking functionality