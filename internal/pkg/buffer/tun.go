package buffer

import (
	"io"
)

func CopyTUN(dst io.Writer, src io.Reader) error {
	bufp := TUNPool.Get().(*[]byte)
	defer TUNPool.Put(bufp)
	buf := *bufp

	_, err := io.CopyBuffer(dst, src, buf)
	return err
}
