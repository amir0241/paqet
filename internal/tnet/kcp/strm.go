package kcp

import (
	"github.com/xtaci/smux"
)

type Strm struct {
	*smux.Stream
	id int
}

func (s *Strm) SID() int {
	return s.id
}
