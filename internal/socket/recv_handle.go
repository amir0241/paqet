package socket

import (
	"context"
	"fmt"
	"net"
	"paqet/internal/conf"
	"runtime"
	"sync"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcap"
)

type recvResult struct {
	payload []byte
	addr    net.Addr
	err     error
}

type RecvHandle struct {
	handle *pcap.Handle
	ch     chan recvResult
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewRecvHandle(cfg *conf.Network) (*RecvHandle, error) {
	handle, err := newHandle(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open pcap handle: %w", err)
	}

	// SetDirection is not fully supported on Windows Npcap, so skip it
	if runtime.GOOS != "windows" {
		if err := handle.SetDirection(pcap.DirectionIn); err != nil {
			handle.Close()
			return nil, fmt.Errorf("failed to set pcap direction in: %v", err)
		}
	}

	filter := fmt.Sprintf("tcp and dst port %d", cfg.Port)
	if err := handle.SetBPFFilter(filter); err != nil {
		handle.Close()
		return nil, fmt.Errorf("failed to set BPF filter: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	h := &RecvHandle{
		handle: handle,
		ch:     make(chan recvResult, 64),
		ctx:    ctx,
		cancel: cancel,
	}

	h.wg.Add(1)
	go h.readLoop()

	return h, nil
}

// readLoop runs in a dedicated goroutine and continuously reads packets from the
// pcap handle, forwarding results on ch. It exits when the handle is closed.
func (h *RecvHandle) readLoop() {
	defer h.wg.Done()
	for {
		payload, addr, err := h.readPacket()
		r := recvResult{payload, addr, err}
		select {
		case h.ch <- r:
		case <-h.ctx.Done():
			return
		}
		if err != nil {
			return
		}
	}
}

func (h *RecvHandle) readPacket() ([]byte, net.Addr, error) {
	data, _, err := h.handle.ReadPacketData()
	if err != nil {
		return nil, nil, err
	}
	p := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)

	addr := &net.UDPAddr{}

	netLayer := p.NetworkLayer()
	if netLayer == nil {
		return nil, nil, nil
	}
	switch netLayer.LayerType() {
	case layers.LayerTypeIPv4:
		addr.IP = netLayer.(*layers.IPv4).SrcIP
	case layers.LayerTypeIPv6:
		addr.IP = netLayer.(*layers.IPv6).SrcIP
	}

	trLayer := p.TransportLayer()
	if trLayer == nil {
		return nil, nil, nil
	}
	switch trLayer.LayerType() {
	case layers.LayerTypeTCP:
		addr.Port = int(trLayer.(*layers.TCP).SrcPort)
	case layers.LayerTypeUDP:
		addr.Port = int(trLayer.(*layers.UDP).SrcPort)
	}

	appLayer := p.ApplicationLayer()
	if appLayer == nil {
		return nil, nil, nil
	}

	return appLayer.Payload(), addr, nil
}

// Read returns the next received packet, blocking until one arrives, the
// provided context is cancelled, or the handle is closed.
func (h *RecvHandle) Read(ctx context.Context) ([]byte, net.Addr, error) {
	select {
	case r := <-h.ch:
		return r.payload, r.addr, r.err
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case <-h.ctx.Done():
		return nil, nil, h.ctx.Err()
	}
}

func (h *RecvHandle) Close() {
	h.cancel()
	if h.handle != nil {
		h.handle.Close()
	}
	h.wg.Wait()
}
