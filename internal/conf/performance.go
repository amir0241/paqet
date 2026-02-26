package conf

import (
	"fmt"
	"paqet/internal/flog"
	"runtime"
)

const (
	// MaxRecommendedConcurrentStreams is the warning threshold for concurrent streams
	MaxRecommendedConcurrentStreams = 100000
)

// Performance configuration for production optimization
type Performance struct {
	// MaxConcurrentStreams limits the number of concurrent stream handlers
	// 0 means unlimited (not recommended for production)
	MaxConcurrentStreams int `yaml:"max_concurrent_streams"`

	// PacketWorkers is the number of parallel packet serialization workers
	// Default is GOMAXPROCS (number of CPU cores)
	PacketWorkers int `yaml:"packet_workers"`

	// StreamWorkerPoolSize is the size of the worker pool for stream handling
	// Default is 1000
	StreamWorkerPoolSize int `yaml:"stream_worker_pool_size"`

	// TCPConnectionPoolSize is the maximum number of cached TCP connections per target
	// 0 disables connection pooling (default)
	TCPConnectionPoolSize int `yaml:"tcp_connection_pool_size"`

	// TCPConnectionIdleTimeout is how long to keep idle TCP connections in seconds
	// Default is 75 seconds
	TCPConnectionIdleTimeout int `yaml:"tcp_connection_idle_timeout"`

	// EnableConnectionPooling enables TCP connection pooling for upstream targets
	EnableConnectionPooling *bool `yaml:"enable_connection_pooling"`

	// MaxRetryAttempts is the maximum number of retry attempts for stream creation
	// Default is 6
	MaxRetryAttempts int `yaml:"max_retry_attempts"`

	// RetryInitialBackoffMs is the initial backoff in milliseconds for retry
	// Default is 100ms
	RetryInitialBackoffMs int `yaml:"retry_initial_backoff_ms"`

	// RetryMaxBackoffMs is the maximum backoff in milliseconds for retry
	// Default is 5000ms (5 seconds)
	RetryMaxBackoffMs int `yaml:"retry_max_backoff_ms"`

	// ConnectionHealthCheckMs controls how often a connection health probe is sent.
	// Lower values detect failures faster but add control-plane overhead.
	ConnectionHealthCheckMs int `yaml:"connection_health_check_ms"`

	// TCPFlagRefreshMs controls how often PTCPF metadata is refreshed to the peer.
	TCPFlagRefreshMs int `yaml:"tcp_flag_refresh_ms"`
}

func (p *Performance) setDefaults(role string) {
	cpus := sysCPUCount()

	if p.MaxConcurrentStreams == 0 {
		// Scale with CPU count to allow more concurrent streams on more capable machines.
		if role == "server" {
			p.MaxConcurrentStreams = clampInt(cpus*12500, 50000, 100000)
		} else {
			p.MaxConcurrentStreams = clampInt(cpus*2500, 10000, 50000)
		}
	}

	if p.PacketWorkers == 0 {
		// Default to number of logical CPUs for optimal parallelism.
		// Clamped to the validation-allowed maximum of 64.
		p.PacketWorkers = clampInt(runtime.GOMAXPROCS(0), 2, 64)
		if role == "server" && p.PacketWorkers < 4 {
			p.PacketWorkers = 4
		}
	}

	if p.StreamWorkerPoolSize == 0 {
		// Scale with CPU count: 2500 per core (server) / 1250 per core (client).
		if role == "server" {
			p.StreamWorkerPoolSize = clampInt(cpus*2500, 10000, 100000)
		} else {
			p.StreamWorkerPoolSize = clampInt(cpus*1250, 5000, 50000)
		}
	}

	if p.TCPConnectionPoolSize == 0 {
		// Scale with CPU count while staying conservative to avoid excess open FDs.
		if role == "server" {
			p.TCPConnectionPoolSize = clampInt(cpus*64, 256, 4096)
		} else {
			p.TCPConnectionPoolSize = clampInt(cpus*16, 64, 512)
		}
	}

	if p.TCPConnectionIdleTimeout == 0 {
		p.TCPConnectionIdleTimeout = 75
	}

	if p.EnableConnectionPooling == nil {
		enabled := role == "server"
		p.EnableConnectionPooling = &enabled
	}

	if p.MaxRetryAttempts == 0 {
		p.MaxRetryAttempts = 6
	}

	if p.RetryInitialBackoffMs == 0 {
		p.RetryInitialBackoffMs = 100
	}

	if p.RetryMaxBackoffMs == 0 {
		p.RetryMaxBackoffMs = 5000
	}

	if p.ConnectionHealthCheckMs == 0 {
		p.ConnectionHealthCheckMs = 1000
	}

	if p.TCPFlagRefreshMs == 0 {
		p.TCPFlagRefreshMs = 5000
	}
}

func (p *Performance) validate() []error {
	var errors []error

	if p.MaxConcurrentStreams < 0 {
		errors = append(errors, fmt.Errorf("max_concurrent_streams must be >= 0 (0 means unlimited)"))
	}

	if p.MaxConcurrentStreams > MaxRecommendedConcurrentStreams {
		flog.Warnf("max_concurrent_streams is very high (%d) - this may cause resource exhaustion", p.MaxConcurrentStreams)
	}

	if p.PacketWorkers < 1 || p.PacketWorkers > 64 {
		errors = append(errors, fmt.Errorf("packet_workers must be between 1 and 64"))
	}

	if p.StreamWorkerPoolSize < 10 || p.StreamWorkerPoolSize > 100000 {
		errors = append(errors, fmt.Errorf("stream_worker_pool_size must be between 10 and 100000"))
	}

	if p.TCPConnectionPoolSize < 0 || p.TCPConnectionPoolSize > 10000 {
		errors = append(errors, fmt.Errorf("tcp_connection_pool_size must be between 0 and 10000"))
	}

	if p.TCPConnectionIdleTimeout < 10 || p.TCPConnectionIdleTimeout > 3600 {
		errors = append(errors, fmt.Errorf("tcp_connection_idle_timeout must be between 10 and 3600 seconds"))
	}

	if p.MaxRetryAttempts < 0 || p.MaxRetryAttempts > 20 {
		errors = append(errors, fmt.Errorf("max_retry_attempts must be between 0 and 20"))
	}

	if p.RetryInitialBackoffMs < 10 || p.RetryInitialBackoffMs > 10000 {
		errors = append(errors, fmt.Errorf("retry_initial_backoff_ms must be between 10 and 10000"))
	}

	if p.RetryMaxBackoffMs < p.RetryInitialBackoffMs || p.RetryMaxBackoffMs > 60000 {
		errors = append(errors, fmt.Errorf("retry_max_backoff_ms must be between retry_initial_backoff_ms and 60000"))
	}

	if p.ConnectionHealthCheckMs < 100 || p.ConnectionHealthCheckMs > 60000 {
		errors = append(errors, fmt.Errorf("connection_health_check_ms must be between 100 and 60000"))
	}

	if p.TCPFlagRefreshMs < 500 || p.TCPFlagRefreshMs > 600000 {
		errors = append(errors, fmt.Errorf("tcp_flag_refresh_ms must be between 500 and 600000"))
	}

	return errors
}

func (p *Performance) ConnectionPoolingEnabled() bool {
	if p.EnableConnectionPooling == nil {
		return false
	}
	return *p.EnableConnectionPooling
}
