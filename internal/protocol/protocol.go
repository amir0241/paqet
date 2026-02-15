package protocol

import (
	"encoding/gob"
	"io"
	"paqet/internal/conf"
	"paqet/internal/tnet"
)

type PType = byte

const (
	PPING PType = 0x01
	PPONG PType = 0x02
	PTCPF PType = 0x03
	PTCP  PType = 0x04
	PUDP  PType = 0x05
	PTUN  PType = 0x06
)

type Proto struct {
	Type PType
	Addr *tnet.Addr
	TCPF []conf.TCPF
}

func (p *Proto) Read(r io.Reader) error {
	dec := gob.NewDecoder(r)

	err := dec.Decode(p)
	if err != nil {
		return err
	}
	return nil
}

func (p *Proto) Write(w io.Writer) error {
	enc := gob.NewEncoder(w)

	err := enc.Encode(p)
	if err != nil {
		return err
	}

	return nil
}

// Send is a helper function to send a protocol message
func Send(w io.Writer, ptype PType, data []byte) error {
	addr, err := tnet.NewAddr(string(data))
	if err != nil {
		// If data is not a valid address, use nil
		addr = nil
	}
	
	p := &Proto{
		Type: ptype,
		Addr: addr,
	}
	return p.Write(w)
}

// TypeTUN is an alias for PTUN for convenience
var TypeTUN = PTUN
