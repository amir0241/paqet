package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"paqet/internal/conf"
	"paqet/internal/flog"
	"paqet/internal/pkg/connpool"
	"paqet/internal/socket"
	"paqet/internal/tnet"
	"paqet/internal/tnet/kcp"
	"paqet/internal/tnet/quic"
)

type Server struct {
	cfg             *conf.Conf
	pConn           *socket.PacketConn
	wg              sync.WaitGroup
	streamSemaphore chan struct{} // Limits concurrent stream processing
	connPools       map[string]*connPoolEntry
	connPoolsMu     sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

type connPoolEntry struct {
	pool       *connpool.ConnPool
	lastAccess time.Time
}

func New(cfg *conf.Conf) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Server{
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize semaphore for limiting concurrent streams
	maxStreams := cfg.Performance.MaxConcurrentStreams
	if maxStreams > 0 {
		s.streamSemaphore = make(chan struct{}, maxStreams)
	}

	// Initialize connection pools map if enabled
	if cfg.Performance.EnableConnectionPooling {
		s.connPools = make(map[string]*connPoolEntry)
	}

	return s, nil
}

// getConnPool gets or creates a connection pool for a specific target address
func (s *Server) getConnPool(addr string) (*connpool.ConnPool, error) {
	if !s.cfg.Performance.EnableConnectionPooling {
		return nil, nil
	}

	s.connPoolsMu.RLock()
	entry, exists := s.connPools[addr]
	s.connPoolsMu.RUnlock()

	if exists {
		// Update last access time
		s.connPoolsMu.Lock()
		entry.lastAccess = time.Now()
		s.connPoolsMu.Unlock()
		return entry.pool, nil
	}

	// Create new pool
	s.connPoolsMu.Lock()
	defer s.connPoolsMu.Unlock()

	// Double-check after acquiring write lock
	entry, exists = s.connPools[addr]
	if exists {
		entry.lastAccess = time.Now()
		return entry.pool, nil
	}

	// Create connection factory
	factory := func(ctx context.Context) (net.Conn, error) {
		dialer := &net.Dialer{Timeout: 10 * time.Second}
		return dialer.DialContext(ctx, "tcp", addr)
	}

	pool, err := connpool.New(
		s.cfg.Performance.TCPConnectionPoolSize,
		time.Duration(s.cfg.Performance.TCPConnectionIdleTimeout)*time.Second,
		factory,
	)
	if err != nil {
		return nil, err
	}

	s.connPools[addr] = &connPoolEntry{
		pool:       pool,
		lastAccess: time.Now(),
	}
	return pool, nil
}

func (s *Server) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		flog.Infof("Shutdown signal received, initiating graceful shutdown...")
		cancel()
	}()

	pConn, err := socket.New(ctx, &s.cfg.Network)
	if err != nil {
		return fmt.Errorf("could not create raw packet conn: %w", err)
	}
	s.pConn = pConn

	var listener tnet.Listener
	switch s.cfg.Transport.Protocol {
	case "kcp":
		listener, err = kcp.Listen(s.cfg.Transport.KCP, pConn)
		if err != nil {
			return fmt.Errorf("could not start KCP listener: %w", err)
		}
	case "quic":
		listener, err = quic.Listen(s.cfg.Transport.QUIC, pConn)
		if err != nil {
			return fmt.Errorf("could not start QUIC listener: %w", err)
		}
		// Set context on QUIC listener for proper cancellation
		if quicListener, ok := listener.(interface{ SetContext(context.Context) }); ok {
			quicListener.SetContext(ctx)
		}
	default:
		return fmt.Errorf("unsupported transport protocol: %s", s.cfg.Transport.Protocol)
	}
	defer listener.Close()

	poolingStatus := "disabled"
	if s.cfg.Performance.EnableConnectionPooling {
		poolingStatus = fmt.Sprintf("enabled (pool size: %d, idle timeout: %ds)",
			s.cfg.Performance.TCPConnectionPoolSize,
			s.cfg.Performance.TCPConnectionIdleTimeout)
		// Start periodic cleanup of unused connection pools
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.cleanupConnPools(ctx)
		}()
	}
	flog.Infof("Server started - listening for packets on :%d (protocol: %s, max concurrent streams: %d, connection pooling: %s)",
		s.cfg.Listen.Addr.Port,
		s.cfg.Transport.Protocol,
		s.cfg.Performance.MaxConcurrentStreams,
		poolingStatus)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.listen(ctx, listener)
	}()

	s.wg.Wait()

	// Close all connection pools
	if s.cfg.Performance.EnableConnectionPooling {
		s.connPoolsMu.Lock()
		for addr, entry := range s.connPools {
			flog.Debugf("closing connection pool for %s", addr)
			entry.pool.Close()
		}
		s.connPoolsMu.Unlock()
	}

	flog.Infof("Server shutdown completed")
	return nil
}

func (s *Server) listen(ctx context.Context, listener tnet.Listener) {
	// Remove the goroutine that causes potential leak
	// The listener's Accept will now handle context cancellation internally
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		conn, err := listener.Accept()
		if err != nil {
			// Check if this is due to context cancellation
			select {
			case <-ctx.Done():
				flog.Debugf("listener accept loop stopped due to context cancellation")
				return
			default:
			}
			flog.Errorf("failed to accept connection: %v", err)
			continue
		}
		flog.Infof("accepted new connection from %s (local: %s)", conn.RemoteAddr(), conn.LocalAddr())

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			defer conn.Close()
			s.handleConn(ctx, conn)
		}()
	}
}

// cleanupConnPools periodically removes unused connection pools to prevent memory leaks
func (s *Server) cleanupConnPools(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	// Pools idle for more than 30 minutes will be removed
	const maxIdleTime = 30 * time.Minute
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.connPoolsMu.Lock()
			now := time.Now()
			toDelete := make([]string, 0)
			
			for addr, entry := range s.connPools {
				if now.Sub(entry.lastAccess) > maxIdleTime {
					toDelete = append(toDelete, addr)
					entry.pool.Close()
				}
			}
			
			for _, addr := range toDelete {
				delete(s.connPools, addr)
			}
			s.connPoolsMu.Unlock()
			
			if len(toDelete) > 0 {
				flog.Debugf("cleaned up %d unused connection pools", len(toDelete))
			}
		}
	}
}
