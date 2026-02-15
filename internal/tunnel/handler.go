package tunnel

import (
	"context"
	"fmt"
	"io"
	"paqet/internal/flog"
	"paqet/internal/pkg/buffer"
	"paqet/internal/tnet"
)

// Handler manages TUN tunnel connections
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

// Start begins handling TUN traffic by creating a stream to the server
func (h *Handler) Start(ctx context.Context) error {
	flog.Infof("Starting TUN tunnel handler for %s", h.tun.Name())

	// Create a TUN stream
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
