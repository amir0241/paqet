package tcp

import (
	"github.com/xtaci/smux"
)

// Strm wraps a smux stream to implement tnet.Strm interface
type Strm struct {
	*smux.Stream
}

// SID returns the stream ID
func (s *Strm) SID() int {
	return int(s.ID())
}
