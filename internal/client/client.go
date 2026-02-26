package client

import (
	"context"
	"paqet/internal/conf"
	"paqet/internal/flog"
	"paqet/internal/pkg/iterator"
	"paqet/internal/tnet"
	"sync"
	"time"
)

type Client struct {
	cfg     *conf.Conf
	iter    *iterator.Iterator[*timedConn]
	udpPool *udpPool
	mu      sync.Mutex
}

func New(cfg *conf.Conf) (*Client, error) {
	c := &Client{
		cfg:     cfg,
		iter:    &iterator.Iterator[*timedConn]{},
		udpPool: &udpPool{strms: make(map[uint64]tnet.Strm)},
	}
	return c, nil
}

func (c *Client) Start(ctx context.Context) error {
	for i := range c.cfg.Transport.Conn {
		tc, err := newTimedConn(ctx, c.cfg)
		if err != nil {
			flog.Errorf("failed to create connection %d: %v", i+1, err)
			return err
		}
		flog.Debugf("client connection %d created successfully", i+1)
		c.iter.Items = append(c.iter.Items, tc)
	}
	// Note: ticker() is currently disabled but kept for potential future use
	// go c.ticker(ctx)
	go c.monitorTransportStats(ctx)

	go func() {
		<-ctx.Done()
		for _, tc := range c.iter.Items {
			tc.close()
		}
		flog.Infof("client shutdown complete")
	}()

	ipv4Addr := "<nil>"
	ipv6Addr := "<nil>"
	if c.cfg.Network.IPv4.Addr != nil {
		ipv4Addr = c.cfg.Network.IPv4.Addr.IP.String()
	}
	if c.cfg.Network.IPv6.Addr != nil {
		ipv6Addr = c.cfg.Network.IPv6.Addr.IP.String()
	}
	flog.Infof("Client started: IPv4:%s IPv6:%s -> %s (%d connections)", ipv4Addr, ipv6Addr, c.cfg.Server.Addr, len(c.iter.Items))
	return nil
}

func (c *Client) monitorTransportStats(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	var lastDropped uint64
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var dropped uint64
			var queueDepth int
			for _, tc := range c.iter.Items {
				if tc == nil || tc.conn == nil {
					continue
				}
				if stats, ok := tc.conn.(interface {
					PacketStats() (uint64, int)
				}); ok {
					d, q := stats.PacketStats()
					dropped += d
					queueDepth += q
				}
			}

			if dropped > lastDropped || queueDepth > 0 {
				flog.Warnf("client packet pressure: dropped=%d (+%d), queue_depth=%d",
					dropped, dropped-lastDropped, queueDepth)
			}
			lastDropped = dropped
		}
	}
}
