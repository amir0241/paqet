package client

import (
	"paqet/internal/flog"
	"paqet/internal/protocol"
	"paqet/internal/tnet"
)

func (c *Client) TUN() (tnet.Strm, error) {
	strm, err := c.newStrm()
	if err != nil {
		flog.Debugf("failed to create stream for TUN: %v", err)
		return nil, err
	}

	p := protocol.Proto{Type: protocol.PTUN, Addr: nil}
	err = p.Write(strm)
	if err != nil {
		flog.Debugf("failed to write TUN protocol header on stream %d: %v", strm.SID(), err)
		strm.Close()
		return nil, err
	}

	flog.Debugf("TUN stream %d created", strm.SID())
	return strm, nil
}
