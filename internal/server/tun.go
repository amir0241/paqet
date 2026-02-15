package server

import (
	"context"
	"io"
	"paqet/internal/flog"
	"paqet/internal/pkg/buffer"
	"paqet/internal/tnet"
)

func (s *Server) handleTUNProtocol(ctx context.Context, strm tnet.Strm) error {
	flog.Infof("TUN stream %d from %s: starting tunnel relay", strm.SID(), strm.RemoteAddr())

	if !s.cfg.TUN.Enabled || s.tun == nil {
		flog.Errorf("TUN stream received but TUN is not enabled on server")
		return io.ErrClosedPipe
	}

	// Start bidirectional relay between stream and TUN device
	errCh := make(chan error, 2)

	// Stream -> TUN (using large buffer pool)
	go func() {
		err := buffer.CopyTUN(ctx, s.tun, strm)
		if err != nil && err != io.EOF && err != context.Canceled {
			flog.Debugf("Stream to TUN copy error: %v", err)
		}
		errCh <- err
	}()

	// TUN -> Stream (using large buffer pool)
	go func() {
		err := buffer.CopyTUN(ctx, strm, s.tun)
		if err != nil && err != io.EOF && err != context.Canceled {
			flog.Debugf("TUN to Stream copy error: %v", err)
		}
		errCh <- err
	}()

	// Wait for error or context cancellation
	select {
	case err := <-errCh:
		if err != context.Canceled && err != io.EOF {
			flog.Infof("TUN stream %d closed with error: %v", strm.SID(), err)
			return err
		}
		flog.Infof("TUN stream %d closed", strm.SID())
		return nil
	case <-ctx.Done():
		flog.Infof("TUN stream %d closed due to context cancellation", strm.SID())
		return ctx.Err()
	}
}
