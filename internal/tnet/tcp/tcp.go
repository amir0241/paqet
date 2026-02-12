package tcp

import (
	"net"
	"paqet/internal/conf"
	"time"

	"github.com/xtaci/smux"
)

// smuxConfig creates a smux configuration based on the TCP transport config
func smuxConfig(cfg *conf.TransportTCP) *smux.Config {
	smuxCfg := smux.DefaultConfig()

	if cfg.SMUXConfig != nil {
		smuxCfg.Version = cfg.SMUXConfig.Version
		smuxCfg.MaxFrameSize = cfg.SMUXConfig.MaxFrameSize
		smuxCfg.MaxReceiveBuffer = cfg.SMUXConfig.MaxReceiveBuffer
		smuxCfg.MaxStreamBuffer = cfg.SMUXConfig.MaxStreamBuffer
		smuxCfg.KeepAliveInterval = time.Duration(cfg.SMUXConfig.KeepAliveInterval) * time.Second
		smuxCfg.KeepAliveTimeout = time.Duration(cfg.SMUXConfig.KeepAliveTimeout) * time.Second
	}

	return smuxCfg
}

// configureTCPConn applies TCP-specific configuration to a connection
func configureTCPConn(conn *net.TCPConn, cfg *conf.TransportTCP) error {
	// Set TCP no delay (disable Nagle's algorithm) for lower latency
	if cfg.NoDelay {
		if err := conn.SetNoDelay(true); err != nil {
			return err
		}
	}

	// Set TCP keep-alive
	if cfg.KeepAlive {
		if err := conn.SetKeepAlive(true); err != nil {
			return err
		}
		if err := conn.SetKeepAlivePeriod(cfg.GetKeepAlivePeriod()); err != nil {
			return err
		}
	}

	// Set read buffer size
	if cfg.ReadBufferSize > 0 {
		if err := conn.SetReadBuffer(cfg.ReadBufferSize); err != nil {
			return err
		}
	}

	// Set write buffer size
	if cfg.WriteBufferSize > 0 {
		if err := conn.SetWriteBuffer(cfg.WriteBufferSize); err != nil {
			return err
		}
	}

	return nil
}
