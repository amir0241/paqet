package server

import (
	"context"
	"io"
	"paqet/internal/flog"
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

	// Stream -> TUN
	go func() {
		buf := make([]byte, s.cfg.TUN.MTU)
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
				n, err := strm.Read(buf)
				if err != nil {
					if err != io.EOF {
						flog.Debugf("TUN stream %d read error: %v", strm.SID(), err)
					}
					errCh <- err
					return
				}
				if n > 0 {
					if _, err := s.tun.Write(buf[:n]); err != nil {
						flog.Debugf("TUN write error: %v", err)
						errCh <- err
						return
					}
					flog.Debugf("Stream %d -> TUN: %d bytes", strm.SID(), n)
				}
			}
		}
	}()

	// TUN -> Stream
	go func() {
		buf := make([]byte, s.cfg.TUN.MTU)
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
				n, err := s.tun.Read(buf)
				if err != nil {
					if err != io.EOF {
						flog.Debugf("TUN read error: %v", err)
					}
					errCh <- err
					return
				}
				if n > 0 {
					if _, err := strm.Write(buf[:n]); err != nil {
						flog.Debugf("TUN stream %d write error: %v", strm.SID(), err)
						errCh <- err
						return
					}
					flog.Debugf("TUN -> Stream %d: %d bytes", strm.SID(), n)
				}
			}
		}
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
