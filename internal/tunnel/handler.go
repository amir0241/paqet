package tunnel

import (
	"context"
	"fmt"
	"io"
	"paqet/internal/flog"
	"paqet/internal/pkg/buffer"
	"paqet/internal/tnet"
)

// Handler manages TUN tunnel connections.
// 
// The Handler creates a private network overlay by establishing a secure tunnel
// between the client and server TUN devices using paqet's encrypted transport (KCP/QUIC).
// All packets sent to the TUN interface are encrypted and transmitted through paqet's
// raw TCP packet transport, creating a VPN-like layer 3 tunnel.
//
// Packet flow:
//   Client TUN device -> Handler -> paqet stream (KCP/QUIC) -> Server -> Server TUN device
//   Server TUN device -> paqet stream (KCP/QUIC) -> Handler -> Client TUN device
type Handler struct {
	tun    *TUN
	client interface {
		TUN() (tnet.Strm, error)
	}
}

// NewHandler creates a new tunnel handler
func NewHandler(tun *TUN, client interface {
	TUN() (tnet.Strm, error)
}) *Handler {
	return &Handler{
		tun:    tun,
		client: client,
	}
}

// Start begins handling TUN traffic by creating a stream to the server.
// 
// This method establishes a secure tunnel through paqet's transport layer:
// 1. Creates a new paqet stream (using KCP or QUIC transport)
// 2. Sends PTUN protocol header to identify this as a TUN stream
// 3. Sets up bidirectional relay between local TUN device and the paqet stream
// 4. All IP packets read from TUN are encrypted and sent through paqet
// 5. All packets received from paqet are decrypted and written to TUN
//
// This creates a private network overlay where traffic between TUN interfaces
// is protected by paqet's encrypted transport, bypassing the host TCP/IP stack.
func (h *Handler) Start(ctx context.Context) error {
	flog.Infof("Starting TUN tunnel handler for %s", h.tun.Name())

	// Create a TUN stream - this establishes a secure paqet connection
	// using the configured transport (KCP or QUIC) with encryption
	strm, err := h.client.TUN()
	if err != nil {
		return fmt.Errorf("failed to create TUN stream: %v", err)
	}
	defer strm.Close()

	flog.Infof("TUN tunnel stream %d established", strm.SID())

	// Start bidirectional copy between TUN device and stream
	errCh := make(chan error, 2)

	// TUN -> Stream (using large buffer pool)
	go func() {
		err := buffer.CopyTUN(ctx, strm, h.tun)
		if err != nil && err != io.EOF && err != context.Canceled {
			flog.Debugf("TUN to Stream copy error: %v", err)
		}
		errCh <- err
	}()

	// Stream -> TUN (using large buffer pool)
	go func() {
		err := buffer.CopyTUN(ctx, h.tun, strm)
		if err != nil && err != io.EOF && err != context.Canceled {
			flog.Debugf("Stream to TUN copy error: %v", err)
		}
		errCh <- err
	}()

	// Wait for error or context cancellation
	select {
	case err := <-errCh:
		if err != context.Canceled && err != io.EOF {
			return fmt.Errorf("tunnel handler error: %v", err)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
