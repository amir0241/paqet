package buffer

import (
	"io"
)

func CopyU(dst io.Writer, src io.Reader) error {
	bufp := UPool.Get()
	defer UPool.Put(bufp)
	buf := *bufp

	_, err := io.CopyBuffer(dst, src, buf)
	return err
}
