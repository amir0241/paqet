package server

import (
	"context"
	"io"
	"paqet/internal/flog"
	"paqet/internal/pkg/buffer"
	"paqet/internal/tnet"
)

// handleTUNProtocol processes TUN tunnel streams from clients.
//
// This method handles the server-side of the TUN tunnel:
// 1. Receives encrypted packets from the paqet stream (sent by client's TUN device)
// 2. Decrypts them (handled by transport layer)
// 3. Writes them to the server's TUN device
// 4. Reads packets from server's TUN device
// 5. Encrypts and sends them back through the paqet stream to the client
//
// This creates a bidirectional encrypted tunnel where IP packets are securely
// relayed between client and server TUN devices through paqet's transport.
func (s *Server) handleTUNProtocol(ctx context.Context, strm tnet.Strm) error {
	flog.Infof("TUN stream %d from %s: starting tunnel relay (packets encrypted via paqet transport)", 
		strm.SID(), strm.RemoteAddr())

	if !s.cfg.TUN.Enabled || s.tun == nil {
		flog.Errorf("TUN stream received but TUN is not enabled on server")
		return io.ErrClosedPipe
	}

	// Start bidirectional relay between paqet stream and TUN device
	// All traffic through this stream is encrypted by the transport layer
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
