package tunnel

import (
	"context"
	"fmt"
	"io"
	"paqet/internal/flog"
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

	// TUN -> Stream
	go func() {
		buf := make([]byte, h.tun.cfg.MTU)
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
				n, err := h.tun.Read(buf)
				if err != nil {
					if err != io.EOF {
						flog.Debugf("TUN read error: %v", err)
					}
					errCh <- err
					return
				}
				if n > 0 {
					if _, err := strm.Write(buf[:n]); err != nil {
						flog.Debugf("Stream write error: %v", err)
						errCh <- err
						return
					}
					flog.Debugf("TUN -> Stream: %d bytes", n)
				}
			}
		}
	}()

	// Stream -> TUN
	go func() {
		buf := make([]byte, h.tun.cfg.MTU)
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
				n, err := strm.Read(buf)
				if err != nil {
					if err != io.EOF {
						flog.Debugf("Stream read error: %v", err)
					}
					errCh <- err
					return
				}
				if n > 0 {
					if _, err := h.tun.Write(buf[:n]); err != nil {
						flog.Debugf("TUN write error: %v", err)
						errCh <- err
						return
					}
					flog.Debugf("Stream -> TUN: %d bytes", n)
				}
			}
		}
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
