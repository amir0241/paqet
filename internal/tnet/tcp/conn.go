package tcp

import (
	"fmt"
	"net"
	"paqet/internal/protocol"
	"paqet/internal/socket"
	"paqet/internal/tnet"
	"time"

	"github.com/xtaci/smux"
)

// Conn wraps a TCP connection with smux to implement tnet.Conn interface
type Conn struct {
	PacketConn *socket.PacketConn
	TCPConn    *net.TCPConn
	Session    *smux.Session
}

// OpenStrm opens a new stream on the smux session
func (c *Conn) OpenStrm() (tnet.Strm, error) {
	strm, err := c.Session.OpenStream()
	if err != nil {
		return nil, err
	}
	return &Strm{strm}, nil
}

// AcceptStrm accepts a new stream from the smux session
func (c *Conn) AcceptStrm() (tnet.Strm, error) {
	strm, err := c.Session.AcceptStream()
	if err != nil {
		return nil, err
	}
	return &Strm{strm}, nil
}

// Ping tests the connection by opening a stream and optionally waiting for a response
func (c *Conn) Ping(wait bool) error {
	strm, err := c.Session.OpenStream()
	if err != nil {
		return fmt.Errorf("ping failed: %v", err)
	}
	defer strm.Close()

	if wait {
		p := protocol.Proto{Type: protocol.PPING}
		err = p.Write(strm)
		if err != nil {
			return fmt.Errorf("stream ping write failed: %v", err)
		}
		err = p.Read(strm)
		if err != nil {
			return fmt.Errorf("stream ping read failed: %v", err)
		}
		if p.Type != protocol.PPONG {
			return fmt.Errorf("stream pong failed: invalid response type")
		}
	}
	return nil
}

// Close closes the smux session, TCP connection, and packet connection
func (c *Conn) Close() error {
	var firstErr error

	if c.Session != nil {
		if err := c.Session.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if c.TCPConn != nil {
		if err := c.TCPConn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if c.PacketConn != nil {
		if err := c.PacketConn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// LocalAddr returns the local network address
func (c *Conn) LocalAddr() net.Addr {
	return c.Session.LocalAddr()
}

// RemoteAddr returns the remote network address
func (c *Conn) RemoteAddr() net.Addr {
	return c.Session.RemoteAddr()
}

// SetDeadline sets the read and write deadlines for the smux session
func (c *Conn) SetDeadline(t time.Time) error {
	return c.Session.SetDeadline(t)
}

// SetReadDeadline sets the read deadline for the TCP connection
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.TCPConn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline for the TCP connection
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.TCPConn.SetWriteDeadline(t)
}
