package client

import (
	"context"
	"paqet/internal/flog"
	"time"
)

func (c *Client) ticker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.reconnectDeadConns(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// reconnectDeadConns checks all connections and proactively reconnects any that
// smux has marked as closed.  This ensures the client recovers even when no
// new incoming connections arrive to trigger the per-request reconnect path.
func (c *Client) reconnectDeadConns(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, tc := range c.iter.Items {
		if tc.conn == nil {
			continue
		}
		if err := tc.conn.Ping(false); err == nil {
			continue // connection still alive
		}
		flog.Infof("proactive reconnect: dead connection detected, reconnecting to %s", c.cfg.Server.Addr)
		tc.conn.Close()
		newConn, err := tc.createConn()
		if err != nil {
			flog.Errorf("proactive reconnect to %s failed: %v", c.cfg.Server.Addr, err)
			continue
		}
		tc.conn = newConn
		flog.Infof("proactive reconnect to %s succeeded", c.cfg.Server.Addr)
	}
}
