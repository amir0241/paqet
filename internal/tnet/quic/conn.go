package quic

import (
	"context"
	"net"
	"paqet/internal/socket"
	"paqet/internal/tnet"
	"time"

	"github.com/quic-go/quic-go"
)

// Conn wraps a QUIC connection to implement the tnet.Conn interface
type Conn struct {
	connection *quic.Conn
	packetConn *socket.PacketConn
	ctx        context.Context
	cancel     context.CancelFunc
}

func newConn(qconn *quic.Conn, pConn *socket.PacketConn) *Conn {
	ctx, cancel := context.WithCancel(context.Background())
	return &Conn{
		connection: qconn,
		packetConn: pConn,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// newConnWithContext creates a Conn with a parent context for proper cancellation propagation
func newConnWithContext(qconn *quic.Conn, pConn *socket.PacketConn, parentCtx context.Context) *Conn {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Conn{
		connection: qconn,
		packetConn: pConn,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (c *Conn) OpenStrm() (tnet.Strm, error) {
	// Add timeout to prevent indefinite blocking under high load
	ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
	defer cancel()

	stream, err := c.connection.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	return &Strm{stream: stream}, nil
}

func (c *Conn) AcceptStrm() (tnet.Strm, error) {
	// Use connection's context which will be cancelled on shutdown
	stream, err := c.connection.AcceptStream(c.ctx)
	if err != nil {
		return nil, err
	}
	return &Strm{stream: stream}, nil
}

func (c *Conn) Ping(wait bool) error {
	// QUIC has built-in keep-alive mechanism
	// We can send a PING frame by trying to open and close a stream
	if wait {
		// Add timeout to prevent indefinite blocking
		ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
		defer cancel()

		stream, err := c.connection.OpenStreamSync(ctx)
		if err != nil {
			return err
		}
		return stream.Close()
	}
	// Non-blocking ping - check connection status
	// Use our context to properly detect shutdown
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
	}

	// Also check the QUIC connection status
	select {
	case <-c.connection.Context().Done():
		return c.connection.Context().Err()
	default:
		return nil
	}
}

func (c *Conn) Close() error {
	c.cancel()
	err := c.connection.CloseWithError(0, "connection closed")
	if c.packetConn != nil {
		if pErr := c.packetConn.Close(); err == nil {
			err = pErr
		}
	}
	return err
}

func (c *Conn) LocalAddr() net.Addr {
	return c.connection.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.connection.RemoteAddr()
}

func (c *Conn) SetDeadline(t time.Time) error {
	// QUIC connections don't support connection-level deadlines
	// Deadlines must be set per-stream using stream.SetDeadline()
	return nil
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	// QUIC connections don't support connection-level deadlines
	// Deadlines must be set per-stream using stream.SetReadDeadline()
	return nil
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	// QUIC connections don't support connection-level deadlines
	// Deadlines must be set per-stream using stream.SetWriteDeadline()
	return nil
}

func (c *Conn) PacketStats() (dropped uint64, queueDepth int) {
	if c.packetConn == nil {
		return 0, 0
	}
	return c.packetConn.DroppedPackets(), c.packetConn.QueueDepth()
}
