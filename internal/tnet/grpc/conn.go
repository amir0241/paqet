package grpc

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"paqet/internal/socket"
	"paqet/internal/tnet"
	pb "paqet/internal/tnet/grpc/proto"

	"google.golang.org/grpc"
)

// Conn wraps a gRPC connection to implement tnet.Conn interface
type Conn struct {
	PacketConn   *socket.PacketConn
	GRPCConn     *grpc.ClientConn
	Client       pb.PaqetTransportClient
	streamClient pb.PaqetTransport_StreamClient
	
	// For server-side connections
	serverStream pb.PaqetTransport_StreamServer
	isServer     bool
	
	localAddr    net.Addr
	remoteAddr   net.Addr
	
	// Stream management
	streamMu      sync.Mutex
	nextStreamID  int32
	activeStreams map[int32]*Strm
	acceptChan    chan *Strm
	
	// Connection state
	closed atomic.Bool
	ctx    context.Context
	cancel context.CancelFunc
}

// NewClientConn creates a new client-side gRPC connection
func NewClientConn(grpcConn *grpc.ClientConn, pConn *socket.PacketConn, remoteAddr net.Addr) (*Conn, error) {
	client := pb.NewPaqetTransportClient(grpcConn)
	
	ctx, cancel := context.WithCancel(context.Background())
	
	// Establish bidirectional stream
	streamClient, err := client.Stream(ctx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}
	
	conn := &Conn{
		PacketConn:    pConn,
		GRPCConn:      grpcConn,
		Client:        client,
		streamClient:  streamClient,
		isServer:      false,
		localAddr:     &net.TCPAddr{IP: net.IPv4zero, Port: 0},
		remoteAddr:    remoteAddr,
		activeStreams: make(map[int32]*Strm),
		acceptChan:    make(chan *Strm, 100),
		ctx:           ctx,
		cancel:        cancel,
	}
	
	// Start receiving streams
	go conn.receiveLoop()
	
	return conn, nil
}

// NewServerConn creates a new server-side gRPC connection
func NewServerConn(serverStream pb.PaqetTransport_StreamServer, remoteAddr net.Addr) (*Conn, error) {
	ctx, cancel := context.WithCancel(serverStream.Context())
	
	conn := &Conn{
		serverStream:  serverStream,
		isServer:      true,
		localAddr:     nil, // Will be set by listener
		remoteAddr:    remoteAddr,
		activeStreams: make(map[int32]*Strm),
		acceptChan:    make(chan *Strm, 100),
		ctx:           ctx,
		cancel:        cancel,
	}
	
	// Start receiving streams
	go conn.receiveLoop()
	
	return conn, nil
}

func (c *Conn) receiveLoop() {
	for !c.closed.Load() {
		var msg *pb.StreamData
		var err error
		
		if c.isServer {
			msg, err = c.serverStream.Recv()
		} else {
			msg, err = c.streamClient.Recv()
		}
		
		if err != nil {
			if err == io.EOF || c.closed.Load() {
				return
			}
			// Handle error
			c.Close()
			return
		}
		
		c.streamMu.Lock()
		strm, exists := c.activeStreams[msg.StreamId]
		if !exists {
			// New incoming stream
			strm = &Strm{
				conn:     c,
				streamID: msg.StreamId,
				recvChan: make(chan []byte, 100),
				closed:   atomic.Bool{},
			}
			c.activeStreams[msg.StreamId] = strm
			
			// Send to accept channel
			select {
			case c.acceptChan <- strm:
			default:
				// Channel full, drop the stream
			}
		}
		c.streamMu.Unlock()
		
		if msg.Close {
			strm.closed.Store(true)
			close(strm.recvChan)
		} else if len(msg.Data) > 0 {
			select {
			case strm.recvChan <- msg.Data:
			default:
				// Channel full, drop the data
			}
		}
	}
}

// OpenStrm opens a new stream
func (c *Conn) OpenStrm() (tnet.Strm, error) {
	if c.closed.Load() {
		return nil, fmt.Errorf("connection closed")
	}
	
	c.streamMu.Lock()
	streamID := atomic.AddInt32(&c.nextStreamID, 1)
	strm := &Strm{
		conn:     c,
		streamID: streamID,
		recvChan: make(chan []byte, 100),
		closed:   atomic.Bool{},
	}
	c.activeStreams[streamID] = strm
	c.streamMu.Unlock()
	
	return strm, nil
}

// AcceptStrm accepts a new stream
func (c *Conn) AcceptStrm() (tnet.Strm, error) {
	select {
	case strm := <-c.acceptChan:
		return strm, nil
	case <-c.ctx.Done():
		return nil, fmt.Errorf("connection closed")
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("accept timeout")
	}
}

// Ping tests the connection
func (c *Conn) Ping(wait bool) error {
	if !wait {
		// Just test if we can open a stream
		strm, err := c.OpenStrm()
		if err != nil {
			return fmt.Errorf("ping failed: %v", err)
		}
		strm.Close()
		return nil
	}
	
	// Send a proper ping request
	if c.isServer {
		return fmt.Errorf("server cannot initiate ping")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	resp, err := c.Client.Ping(ctx, &pb.PingRequest{
		Timestamp: time.Now().Unix(),
	})
	if err != nil {
		return fmt.Errorf("ping failed: %v", err)
	}
	
	if resp.Timestamp == 0 {
		return fmt.Errorf("invalid pong response")
	}
	
	return nil
}

// Close closes the connection
func (c *Conn) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return nil // Already closed
	}
	
	c.cancel()
	
	var firstErr error
	
	// Close all active streams
	c.streamMu.Lock()
	for _, strm := range c.activeStreams {
		if err := strm.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	c.streamMu.Unlock()
	
	if !c.isServer && c.GRPCConn != nil {
		if err := c.GRPCConn.Close(); err != nil && firstErr == nil {
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
	if c.localAddr != nil {
		return c.localAddr
	}
	return &net.TCPAddr{IP: net.IPv4zero, Port: 0}
}

// RemoteAddr returns the remote network address
func (c *Conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

// SetDeadline sets deadlines (not supported for gRPC)
func (c *Conn) SetDeadline(t time.Time) error {
	// gRPC handles deadlines via context
	return nil
}

// SetReadDeadline sets read deadline (not supported for gRPC)
func (c *Conn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets write deadline (not supported for gRPC)
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return nil
}

// sendData sends data on a stream
func (c *Conn) sendData(streamID int32, data []byte, close bool) error {
	if c.closed.Load() {
		return fmt.Errorf("connection closed")
	}
	
	msg := &pb.StreamData{
		StreamId: streamID,
		Data:     data,
		Close:    close,
	}
	
	if c.isServer {
		return c.serverStream.Send(msg)
	}
	return c.streamClient.Send(msg)
}
