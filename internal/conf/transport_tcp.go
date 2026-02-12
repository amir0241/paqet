package conf

import (
	"fmt"
	"time"
)

// TransportTCP holds TCP transport-specific configuration
type TransportTCP struct {
	// KeepAlive enables TCP keep-alive probes (default: true)
	KeepAlive bool `yaml:"keep_alive"`

	// KeepAlivePeriod is the interval between keep-alive probes in seconds (default: 30)
	KeepAlivePeriod int `yaml:"keep_alive_period"`

	// NoDelay disables Nagle's algorithm for lower latency (default: true)
	NoDelay bool `yaml:"no_delay"`

	// ReadBufferSize is the size of the TCP read buffer in bytes (default: 4MB)
	ReadBufferSize int `yaml:"read_buffer_size"`

	// WriteBufferSize is the size of the TCP write buffer in bytes (default: 4MB)
	WriteBufferSize int `yaml:"write_buffer_size"`

	// SMUX multiplexing configuration
	SMUXConfig *SMUXConfig `yaml:"smux"`
}

// SMUXConfig holds smux multiplexing settings for TCP
type SMUXConfig struct {
	// Version of smux protocol (default: 1)
	Version int `yaml:"version"`

	// MaxFrameSize is maximum frame size in bytes (default: 32KB)
	MaxFrameSize int `yaml:"max_frame_size"`

	// MaxReceiveBuffer is the maximum receive buffer size (default: 4MB)
	MaxReceiveBuffer int `yaml:"max_receive_buffer"`

	// MaxStreamBuffer is the maximum stream buffer size (default: 2MB)
	MaxStreamBuffer int `yaml:"max_stream_buffer"`

	// KeepAliveInterval is the interval for smux keep-alive in seconds (default: 10)
	KeepAliveInterval int `yaml:"keep_alive_interval"`

	// KeepAliveTimeout is the timeout for smux keep-alive in seconds (default: 30)
	KeepAliveTimeout int `yaml:"keep_alive_timeout"`
}

func (t *TransportTCP) setDefaults(role string) {
	// Note: TCP transport uses the same defaults for both client and server roles
	// Unlike KCP and QUIC which have role-specific optimizations

	// TCP connection settings
	if t.KeepAlivePeriod == 0 {
		t.KeepAlivePeriod = 30
	}

	if t.ReadBufferSize == 0 {
		t.ReadBufferSize = 4 * 1024 * 1024 // 4MB
	}

	if t.WriteBufferSize == 0 {
		t.WriteBufferSize = 4 * 1024 * 1024 // 4MB
	}

	// Initialize SMUX config if not provided
	if t.SMUXConfig == nil {
		t.SMUXConfig = &SMUXConfig{}
	}

	// SMUX defaults
	if t.SMUXConfig.Version == 0 {
		t.SMUXConfig.Version = 1
	}

	if t.SMUXConfig.MaxFrameSize == 0 {
		t.SMUXConfig.MaxFrameSize = 32 * 1024 // 32KB
	}

	if t.SMUXConfig.MaxReceiveBuffer == 0 {
		t.SMUXConfig.MaxReceiveBuffer = 4 * 1024 * 1024 // 4MB
	}

	if t.SMUXConfig.MaxStreamBuffer == 0 {
		t.SMUXConfig.MaxStreamBuffer = 2 * 1024 * 1024 // 2MB
	}

	if t.SMUXConfig.KeepAliveInterval == 0 {
		t.SMUXConfig.KeepAliveInterval = 10
	}

	if t.SMUXConfig.KeepAliveTimeout == 0 {
		t.SMUXConfig.KeepAliveTimeout = 30
	}
}

func (t *TransportTCP) validate() []error {
	var errors []error

	// Validate keep-alive period
	if t.KeepAlivePeriod < 1 || t.KeepAlivePeriod > 7200 {
		errors = append(errors, fmt.Errorf("TCP keep_alive_period must be between 1-7200 seconds"))
	}

	// Validate buffer sizes
	if t.ReadBufferSize < 1024 {
		errors = append(errors, fmt.Errorf("TCP read_buffer_size must be at least 1024 bytes"))
	}

	if t.WriteBufferSize < 1024 {
		errors = append(errors, fmt.Errorf("TCP write_buffer_size must be at least 1024 bytes"))
	}

	// Validate SMUX settings
	if t.SMUXConfig != nil {
		if t.SMUXConfig.Version != 1 && t.SMUXConfig.Version != 2 {
			errors = append(errors, fmt.Errorf("SMUX version must be 1 or 2"))
		}

		if t.SMUXConfig.MaxFrameSize < 1024 || t.SMUXConfig.MaxFrameSize > 1024*1024 {
			errors = append(errors, fmt.Errorf("SMUX max_frame_size must be between 1KB-1MB"))
		}

		if t.SMUXConfig.MaxReceiveBuffer < 1024*1024 {
			errors = append(errors, fmt.Errorf("SMUX max_receive_buffer must be at least 1MB"))
		}

		if t.SMUXConfig.MaxStreamBuffer < 512*1024 {
			errors = append(errors, fmt.Errorf("SMUX max_stream_buffer must be at least 512KB"))
		}

		if t.SMUXConfig.KeepAliveInterval < 1 || t.SMUXConfig.KeepAliveInterval > 300 {
			errors = append(errors, fmt.Errorf("SMUX keep_alive_interval must be between 1-300 seconds"))
		}

		if t.SMUXConfig.KeepAliveTimeout < 1 || t.SMUXConfig.KeepAliveTimeout > 600 {
			errors = append(errors, fmt.Errorf("SMUX keep_alive_timeout must be between 1-600 seconds"))
		}
	}

	return errors
}

// GetKeepAlivePeriod returns the keep-alive period as a time.Duration
func (t *TransportTCP) GetKeepAlivePeriod() time.Duration {
	return time.Duration(t.KeepAlivePeriod) * time.Second
}
