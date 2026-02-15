package client

import (
	"paqet/internal/flog"
	"paqet/internal/protocol"
	"paqet/internal/tnet"
)

// TUN creates a new paqet stream for TUN tunnel traffic.
//
// This method establishes the client-side connection for TUN mode:
// 1. Creates a new encrypted stream using the configured transport (KCP/QUIC)
// 2. Sends PTUN protocol header to inform the server this is a TUN tunnel
// 3. Returns the stream for bidirectional packet relay
//
// All IP packets from the client's TUN device will be sent through this stream,
// encrypted by paqet's transport layer, and relayed to the server's TUN device.
// This creates a secure layer 3 tunnel through paqet's raw packet transport.
func (c *Client) TUN() (tnet.Strm, error) {
	// Create a new paqet stream - this uses KCP or QUIC with encryption
	strm, err := c.newStrm()
	if err != nil {
		flog.Debugf("failed to create stream for TUN: %v", err)
		return nil, err
	}

	// Send TUN protocol header to identify this stream's purpose
	p := protocol.Proto{Type: protocol.PTUN, Addr: nil}
	err = p.Write(strm)
	if err != nil {
		flog.Debugf("failed to write TUN protocol header on stream %d: %v", strm.SID(), err)
		strm.Close()
		return nil, err
	}

	flog.Debugf("TUN stream %d created (traffic will be encrypted via paqet transport)", strm.SID())
	return strm, nil
}
