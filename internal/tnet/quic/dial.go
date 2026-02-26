package quic

import (
	"context"
	"fmt"
	"net"
	"paqet/internal/conf"
	"paqet/internal/flog"
	"paqet/internal/socket"
	"paqet/internal/tnet"
	"time"

	"github.com/quic-go/quic-go"
)

func Dial(addr *net.UDPAddr, cfg *conf.QUIC, pConn *socket.PacketConn) (tnet.Conn, error) {
	// Generate TLS config for client
	tlsConfig, err := cfg.GenerateTLSConfig("client")
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS config: %w", err)
	}

	// Set server name if not already set
	if tlsConfig.ServerName == "" {
		tlsConfig.ServerName = addr.IP.String()
	}

	// Create QUIC config
	quicConfig := getQUICConfig(cfg)

	flog.Debugf("QUIC dialing %s", addr.String())

	// Use context with timeout to prevent indefinite dial attempts
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Dial QUIC connection using the packet connection
	qconn, err := quic.Dial(ctx, pConn, addr, tlsConfig, quicConfig)
	if err != nil {
		return nil, fmt.Errorf("QUIC connection attempt failed: %v", err)
	}

	flog.Debugf("QUIC connection established to %s", addr.String())

	return newConn(qconn, pConn), nil
}
