package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"paqet/internal/conf"
	"paqet/internal/flog"
	"paqet/internal/socket"
	"paqet/internal/tnet"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// Dial creates a gRPC connection to the specified address
func Dial(addr *net.UDPAddr, cfg *conf.GRPC, pConn *socket.PacketConn) (tnet.Conn, error) {
	// Convert UDP address to TCP address
	tcpAddr := &net.TCPAddr{
		IP:   addr.IP,
		Port: addr.Port,
		Zone: addr.Zone,
	}

	flog.Debugf("gRPC dialing %s", tcpAddr.String())

	// Prepare dial options
	var opts []grpc.DialOption

	// TLS configuration
	tlsConfig, err := cfg.GenerateTLSConfig("client")
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS config: %w", err)
	}

	if tlsConfig.InsecureSkipVerify {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	// Apply configuration options
	if cfg.InitialWindowSize > 0 {
		opts = append(opts, grpc.WithInitialWindowSize(cfg.InitialWindowSize))
	}
	if cfg.InitialConnWindowSize > 0 {
		opts = append(opts, grpc.WithInitialConnWindowSize(cfg.InitialConnWindowSize))
	}
	if cfg.WriteBufferSize > 0 {
		opts = append(opts, grpc.WithWriteBufferSize(cfg.WriteBufferSize))
	}
	if cfg.ReadBufferSize > 0 {
		opts = append(opts, grpc.WithReadBufferSize(cfg.ReadBufferSize))
	}

	// Set keep-alive parameters
	opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                time.Duration(cfg.KeepAliveTime) * time.Second,
		Timeout:             time.Duration(cfg.KeepAliveTimeout) * time.Second,
		PermitWithoutStream: true,
	}))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Dial gRPC connection
	grpcConn, err := grpc.DialContext(ctx, tcpAddr.String(), opts...)
	if err != nil {
		return nil, fmt.Errorf("gRPC connection attempt failed: %v", err)
	}

	flog.Debugf("gRPC connection established to %s", tcpAddr.String())

	// Create and return connection
	acceptTimeout := time.Duration(cfg.AcceptTimeout) * time.Second
	conn, err := NewClientConn(grpcConn, pConn, tcpAddr, acceptTimeout)
	if err != nil {
		grpcConn.Close()
		return nil, fmt.Errorf("failed to create client connection: %w", err)
	}

	// Set read timeout on streams
	conn.streamMu.Lock()
	for _, strm := range conn.activeStreams {
		strm.readTimeout = time.Duration(cfg.ReadTimeout) * time.Second
	}
	conn.streamMu.Unlock()

	return conn, nil
}
