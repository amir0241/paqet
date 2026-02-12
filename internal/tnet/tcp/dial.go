package tcp

import (
	"fmt"
	"net"
	"paqet/internal/conf"
	"paqet/internal/flog"
	"paqet/internal/socket"
	"paqet/internal/tnet"
	"time"

	"github.com/xtaci/smux"
)

// Dial creates a TCP connection to the specified address and wraps it with smux
func Dial(addr *net.UDPAddr, cfg *conf.TransportTCP, pConn *socket.PacketConn) (tnet.Conn, error) {
	// Convert UDP address to TCP address (paqet uses UDP addresses for consistency across transports)
	tcpAddr := &net.TCPAddr{
		IP:   addr.IP,
		Port: addr.Port,
		Zone: addr.Zone,
	}

	flog.Debugf("TCP dialing %s", tcpAddr.String())

	// Create TCP dialer with timeout
	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}

	// Dial TCP connection
	conn, err := dialer.Dial("tcp", tcpAddr.String())
	if err != nil {
		return nil, fmt.Errorf("TCP connection attempt failed: %v", err)
	}

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		conn.Close()
		return nil, fmt.Errorf("expected TCP connection, got %T", conn)
	}

	// Apply TCP-specific configuration
	if err := configureTCPConn(tcpConn, cfg); err != nil {
		tcpConn.Close()
		return nil, fmt.Errorf("failed to configure TCP connection: %w", err)
	}

	flog.Debugf("TCP connection created, creating smux session")

	// Create smux client session
	sess, err := smux.Client(tcpConn, smuxConfig(cfg))
	if err != nil {
		tcpConn.Close()
		return nil, fmt.Errorf("failed to create smux session: %w", err)
	}

	flog.Debugf("smux session created successfully")
	return &Conn{
		PacketConn: pConn,
		TCPConn:    tcpConn,
		Session:    sess,
	}, nil
}
