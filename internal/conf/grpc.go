package conf

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

type GRPC struct {
	// Connection settings
	MaxConnectionIdle     int `yaml:"max_connection_idle"`      // Maximum idle time before connection is closed (default: 300s)
	MaxConnectionAge      int `yaml:"max_connection_age"`       // Maximum age of connection (default: unlimited)
	MaxConnectionAgeGrace int `yaml:"max_connection_age_grace"` // Grace period for active RPCs after MaxConnectionAge (default: unlimited)
	KeepAliveTime         int `yaml:"keep_alive_time"`          // Server-side keep-alive interval (default: 7200s)
	KeepAliveTimeout      int `yaml:"keep_alive_timeout"`       // Keep-alive timeout (default: 20s)

	// Stream settings
	MaxConcurrentStreams uint32 `yaml:"max_concurrent_streams"`  // Maximum concurrent streams per connection (default: 100)
	InitialWindowSize    int32  `yaml:"initial_window_size"`     // Initial window size for stream-level flow control (default: 64KB)
	InitialConnWindowSize int32  `yaml:"initial_conn_window_size"` // Initial window size for connection-level flow control (default: 16MB)

	// Buffer settings
	WriteBufferSize int `yaml:"write_buffer_size"` // Size of the write buffer (default: 32KB)
	ReadBufferSize  int `yaml:"read_buffer_size"`  // Size of the read buffer (default: 32KB)

	// Timeout settings
	AcceptTimeout int `yaml:"accept_timeout"` // Timeout for accepting streams (default: 30s)
	ReadTimeout   int `yaml:"read_timeout"`   // Timeout for reading from streams (default: 30s)

	// TLS settings
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"` // Skip TLS verification (default: false)
	ServerName         string `yaml:"server_name"`          // Server name for TLS verification

	// Internal TLS config (not exposed to YAML)
	TLSConfig *tls.Config `yaml:"-"`
}

func (g *GRPC) setDefaults(role string) {
	if g.MaxConnectionIdle == 0 {
		g.MaxConnectionIdle = 300 // 5 minutes
	}

	if g.KeepAliveTime == 0 {
		g.KeepAliveTime = 7200 // 2 hours
	}

	if g.KeepAliveTimeout == 0 {
		g.KeepAliveTimeout = 20 // 20 seconds
	}

	if g.MaxConcurrentStreams == 0 {
		if role == "server" {
			g.MaxConcurrentStreams = 1000
		} else {
			g.MaxConcurrentStreams = 100
		}
	}

	if g.InitialWindowSize == 0 {
		g.InitialWindowSize = 65535 // 64KB (HTTP/2 default)
	}

	if g.InitialConnWindowSize == 0 {
		g.InitialConnWindowSize = 16 * 1024 * 1024 // 16MB
	}

	if g.WriteBufferSize == 0 {
		g.WriteBufferSize = 32 * 1024 // 32KB
	}

	if g.ReadBufferSize == 0 {
		g.ReadBufferSize = 32 * 1024 // 32KB
	}

	if g.AcceptTimeout == 0 {
		g.AcceptTimeout = 30 // 30 seconds
	}

	if g.ReadTimeout == 0 {
		g.ReadTimeout = 30 // 30 seconds
	}
}

func (g *GRPC) validate() []error {
	var errors []error

	if g.MaxConnectionIdle < 0 {
		errors = append(errors, fmt.Errorf("gRPC max_connection_idle must be >= 0"))
	}

	if g.MaxConnectionAge < 0 {
		errors = append(errors, fmt.Errorf("gRPC max_connection_age must be >= 0"))
	}

	if g.MaxConnectionAgeGrace < 0 {
		errors = append(errors, fmt.Errorf("gRPC max_connection_age_grace must be >= 0"))
	}

	if g.KeepAliveTime < 1 || g.KeepAliveTime > 86400 {
		errors = append(errors, fmt.Errorf("gRPC keep_alive_time must be between 1-86400 seconds"))
	}

	if g.KeepAliveTimeout < 1 || g.KeepAliveTimeout > 600 {
		errors = append(errors, fmt.Errorf("gRPC keep_alive_timeout must be between 1-600 seconds"))
	}

	if g.MaxConcurrentStreams < 1 {
		errors = append(errors, fmt.Errorf("gRPC max_concurrent_streams must be >= 1"))
	}

	if g.InitialWindowSize < 65535 {
		errors = append(errors, fmt.Errorf("gRPC initial_window_size must be >= 65535 (64KB)"))
	}

	if g.InitialConnWindowSize < 65535 {
		errors = append(errors, fmt.Errorf("gRPC initial_conn_window_size must be >= 65535 (64KB)"))
	}

	if g.WriteBufferSize < 1024 {
		errors = append(errors, fmt.Errorf("gRPC write_buffer_size must be >= 1024 bytes"))
	}

	if g.ReadBufferSize < 1024 {
		errors = append(errors, fmt.Errorf("gRPC read_buffer_size must be >= 1024 bytes"))
	}

	if g.AcceptTimeout < 1 || g.AcceptTimeout > 300 {
		errors = append(errors, fmt.Errorf("gRPC accept_timeout must be between 1-300 seconds"))
	}

	if g.ReadTimeout < 1 || g.ReadTimeout > 300 {
		errors = append(errors, fmt.Errorf("gRPC read_timeout must be between 1-300 seconds"))
	}

	return errors
}

// GenerateTLSConfig generates a TLS configuration for gRPC
func (g *GRPC) GenerateTLSConfig(role string) (*tls.Config, error) {
	if role == "server" {
		// Generate self-signed certificate for server
		cert, err := generateGRPCSelfSignedCert()
		if err != nil {
			return nil, fmt.Errorf("failed to generate self-signed certificate: %w", err)
		}

		return &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"h2"}, // HTTP/2 for gRPC
			MinVersion:   tls.VersionTLS12,
		}, nil
	}

	// Client configuration
	tlsConfig := &tls.Config{
		NextProtos:         []string{"h2"},
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: g.InsecureSkipVerify,
	}

	if g.ServerName != "" {
		tlsConfig.ServerName = g.ServerName
	}

	return tlsConfig, nil
}

func generateGRPCSelfSignedCert() (tls.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		// Add SAN for modern TLS implementations
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tlsCert, nil
}
