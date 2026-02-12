package grpc

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Strm implements tnet.Strm interface for gRPC streams
type Strm struct {
	conn     *Conn
	streamID int32
	recvChan chan []byte
	recvBuf  []byte
	recvMu   sync.Mutex
	closed   atomic.Bool
}

// Read reads data from the stream
func (s *Strm) Read(b []byte) (int, error) {
	if s.closed.Load() {
		return 0, io.EOF
	}
	
	s.recvMu.Lock()
	defer s.recvMu.Unlock()
	
	// If we have buffered data, use it first
	if len(s.recvBuf) > 0 {
		n := copy(b, s.recvBuf)
		s.recvBuf = s.recvBuf[n:]
		return n, nil
	}
	
	// Wait for new data
	select {
	case data, ok := <-s.recvChan:
		if !ok {
			return 0, io.EOF
		}
		n := copy(b, data)
		if n < len(data) {
			// Save remaining data for next read
			s.recvBuf = data[n:]
		}
		return n, nil
	case <-time.After(30 * time.Second):
		return 0, io.ErrNoProgress
	}
}

// Write writes data to the stream
func (s *Strm) Write(b []byte) (int, error) {
	if s.closed.Load() {
		return 0, io.ErrClosedPipe
	}
	
	// Make a copy of the data to send
	data := make([]byte, len(b))
	copy(data, b)
	
	if err := s.conn.sendData(s.streamID, data, false); err != nil {
		return 0, err
	}
	
	return len(b), nil
}

// Close closes the stream
func (s *Strm) Close() error {
	if !s.closed.CompareAndSwap(false, true) {
		return nil // Already closed
	}
	
	// Send close message
	_ = s.conn.sendData(s.streamID, nil, true)
	
	// Remove from active streams
	s.conn.streamMu.Lock()
	delete(s.conn.activeStreams, s.streamID)
	s.conn.streamMu.Unlock()
	
	return nil
}

// LocalAddr returns the local address
func (s *Strm) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

// RemoteAddr returns the remote address
func (s *Strm) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

// SetDeadline sets deadlines (not fully supported)
func (s *Strm) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets read deadline (not fully supported)
func (s *Strm) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets write deadline (not fully supported)
func (s *Strm) SetWriteDeadline(t time.Time) error {
	return nil
}

// SID returns the stream ID
func (s *Strm) SID() int {
	return int(s.streamID)
}
