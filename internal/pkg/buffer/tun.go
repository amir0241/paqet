package buffer

import (
	"context"
	"io"
)

// contextReader wraps an io.Reader to respect context cancellation
type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (cr *contextReader) Read(p []byte) (int, error) {
	// Check if context is cancelled before reading
	select {
	case <-cr.ctx.Done():
		return 0, cr.ctx.Err()
	default:
	}
	return cr.r.Read(p)
}

// contextWriter wraps an io.Writer to respect context cancellation
type contextWriter struct {
	ctx context.Context
	w   io.Writer
}

func (cw *contextWriter) Write(p []byte) (int, error) {
	// Check if context is cancelled before writing
	select {
	case <-cw.ctx.Done():
		return 0, cw.ctx.Err()
	default:
	}
	return cw.w.Write(p)
}

// CopyTUN copies from src to dst using the TUN buffer pool with context awareness
func CopyTUN(ctx context.Context, dst io.Writer, src io.Reader) error {
	bufp := TUNPool.Get().(*[]byte)
	defer TUNPool.Put(bufp)
	buf := *bufp

	// Wrap readers and writers with context awareness
	ctxSrc := &contextReader{ctx: ctx, r: src}
	ctxDst := &contextWriter{ctx: ctx, w: dst}

	_, err := io.CopyBuffer(ctxDst, ctxSrc, buf)
	return err
}
