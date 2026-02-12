package tcp

import (
	"fmt"
	"net"
	"paqet/internal/conf"
	"paqet/internal/flog"
	"paqet/internal/socket"
	"paqet/internal/tnet"

	"github.com/xtaci/smux"
)

// Listener implements tnet.Listener for TCP connections
type Listener struct {
	packetConn *socket.PacketConn
	cfg        *conf.TransportTCP
	listener   *net.TCPListener
}

// Listen creates a TCP listener that accepts connections and wraps them with smux
func Listen(cfg *conf.TransportTCP, pConn *socket.PacketConn) (tnet.Listener, error) {
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

	flog.Debugf("TCP listening on %s", tcpAddr.String())

	// Create TCP listener
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create TCP listener: %w", err)
	}

	return &Listener{
		packetConn: pConn,
		cfg:        cfg,
		listener:   listener,
	}, nil
}

// Accept accepts a new TCP connection and wraps it with smux
func (l *Listener) Accept() (tnet.Conn, error) {
	conn, err := l.listener.AcceptTCP()
	if err != nil {
		return nil, err
	}

	// Apply TCP-specific configuration
	if err := configureTCPConn(conn, l.cfg); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to configure TCP connection: %w", err)
	}

	// Create smux server session
	sess, err := smux.Server(conn, smuxConfig(l.cfg))
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create smux session: %w", err)
	}

	flog.Debugf("Accepted TCP connection from %s", conn.RemoteAddr())

	return &Conn{
		PacketConn: nil, // Server-side connections don't need packet conn
		TCPConn:    conn,
		Session:    sess,
	}, nil
}

// Close closes the TCP listener and associated packet connection
func (l *Listener) Close() error {
	var firstErr error

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
