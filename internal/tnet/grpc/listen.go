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
	pb "paqet/internal/tnet/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// Listener implements tnet.Listener for gRPC connections
type Listener struct {
	packetConn    *socket.PacketConn
	cfg           *conf.GRPC
	grpcServer    *grpc.Server
	listener      net.Listener
	acceptChan    chan tnet.Conn
	acceptTimeout time.Duration
	readTimeout   time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
}

// Listen creates a gRPC listener
func Listen(cfg *conf.GRPC, pConn *socket.PacketConn) (tnet.Listener, error) {
	// Get the local address from the packet connection
	localAddr := pConn.LocalAddr()

	// Convert to TCP address
	var tcpAddr *net.TCPAddr
	switch addr := localAddr.(type) {
	case *net.UDPAddr:
		tcpAddr = &net.TCPAddr{
			IP:   addr.IP,
			Port: addr.Port,
			Zone: addr.Zone,
		}
	case *net.TCPAddr:
		tcpAddr = addr
	default:
		return nil, fmt.Errorf("unsupported address type: %T", localAddr)
	}

	flog.Debugf("gRPC listening on %s", tcpAddr.String())

	// Create TCP listener
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create TCP listener: %w", err)
	}

	// Prepare server options
	var opts []grpc.ServerOption

	// TLS configuration
	tlsConfig, err := cfg.GenerateTLSConfig("server")
	if err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to generate TLS config: %w", err)
	}
	opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))

	// Apply configuration options
	if cfg.MaxConcurrentStreams > 0 {
		opts = append(opts, grpc.MaxConcurrentStreams(cfg.MaxConcurrentStreams))
	}
	if cfg.InitialWindowSize > 0 {
		opts = append(opts, grpc.InitialWindowSize(cfg.InitialWindowSize))
	}
	if cfg.InitialConnWindowSize > 0 {
		opts = append(opts, grpc.InitialConnWindowSize(cfg.InitialConnWindowSize))
	}
	if cfg.WriteBufferSize > 0 {
		opts = append(opts, grpc.WriteBufferSize(cfg.WriteBufferSize))
	}
	if cfg.ReadBufferSize > 0 {
		opts = append(opts, grpc.ReadBufferSize(cfg.ReadBufferSize))
	}

	// Set keep-alive parameters
	kaep := keepalive.EnforcementPolicy{
		MinTime:             time.Duration(cfg.KeepAliveTime) * time.Second,
		PermitWithoutStream: true,
	}
	kasp := keepalive.ServerParameters{
		MaxConnectionIdle:     time.Duration(cfg.MaxConnectionIdle) * time.Second,
		MaxConnectionAge:      time.Duration(cfg.MaxConnectionAge) * time.Second,
		MaxConnectionAgeGrace: time.Duration(cfg.MaxConnectionAgeGrace) * time.Second,
		Time:                  time.Duration(cfg.KeepAliveTime) * time.Second,
		Timeout:               time.Duration(cfg.KeepAliveTimeout) * time.Second,
	}
	opts = append(opts, grpc.KeepaliveEnforcementPolicy(kaep))
	opts = append(opts, grpc.KeepaliveParams(kasp))

	// Create gRPC server
	grpcServer := grpc.NewServer(opts...)

	ctx, cancel := context.WithCancel(context.Background())

	acceptTimeout := time.Duration(cfg.AcceptTimeout) * time.Second
	readTimeout := time.Duration(cfg.ReadTimeout) * time.Second

	l := &Listener{
		packetConn:    pConn,
		cfg:           cfg,
		grpcServer:    grpcServer,
		listener:      listener,
		acceptChan:    make(chan tnet.Conn, 10),
		acceptTimeout: acceptTimeout,
		readTimeout:   readTimeout,
		ctx:           ctx,
		cancel:        cancel,
	}

	// Register the transport service
	pb.RegisterPaqetTransportServer(grpcServer, &transportServer{listener: l})

	// Start serving in a goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			flog.Errorf("gRPC server error: %v", err)
		}
	}()

	return l, nil
}

// Accept accepts a new gRPC connection
func (l *Listener) Accept() (tnet.Conn, error) {
	select {
	case conn := <-l.acceptChan:
		return conn, nil
	case <-l.ctx.Done():
		return nil, fmt.Errorf("listener closed")
	}
}

// Close closes the gRPC listener
func (l *Listener) Close() error {
	l.cancel()

	var firstErr error

	if l.grpcServer != nil {
		l.grpcServer.GracefulStop()
	}

	if l.listener != nil {
		if err := l.listener.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if l.packetConn != nil {
		if err := l.packetConn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// Addr returns the listener's network address
func (l *Listener) Addr() net.Addr {
	return l.listener.Addr()
}

// transportServer implements the PaqetTransportServer interface
type transportServer struct {
	pb.UnimplementedPaqetTransportServer
	listener *Listener
}

func (s *transportServer) Stream(stream pb.PaqetTransport_StreamServer) error {
	// Get remote address from stream context
	remoteAddr := &net.TCPAddr{IP: net.IPv4zero, Port: 0}

	// Create server connection with timeouts
	conn, err := NewServerConn(stream, remoteAddr, s.listener.acceptTimeout)
	if err != nil {
		return fmt.Errorf("failed to create server connection: %w", err)
	}

	// Set read timeout on future streams
	conn.streamMu.Lock()
	for _, strm := range conn.activeStreams {
		strm.readTimeout = s.listener.readTimeout
	}
	conn.streamMu.Unlock()

	// Send connection to accept channel
	select {
	case s.listener.acceptChan <- conn:
	case <-s.listener.ctx.Done():
		conn.Close()
		return fmt.Errorf("listener closed")
	case <-time.After(s.listener.acceptTimeout):
		conn.Close()
		return fmt.Errorf("accept timeout")
	}

	// Wait for stream to close
	<-stream.Context().Done()
	return nil
}

func (s *transportServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PongResponse, error) {
	return &pb.PongResponse{
		Timestamp: time.Now().Unix(),
	}, nil
}
