package tnet

import (
	"net"
	"sync/atomic"
)

var globalSID atomic.Int64

// NextSID returns a globally unique stream ID.
func NextSID() int {
	return int(globalSID.Add(1))
}

type Strm interface {
	net.Conn
	SID() int
}
