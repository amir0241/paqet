package client

import (
	"fmt"
	"math"
	"paqet/internal/flog"
	"paqet/internal/tnet"
	"time"
)

func (c *Client) newConn(forceCheck bool) (tnet.Conn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	tc := c.iter.Next()
	if tc == nil {
		return nil, fmt.Errorf("no available connections")
	}

	healthEvery := time.Duration(c.cfg.Performance.ConnectionHealthCheckMs) * time.Millisecond
	if healthEvery <= 0 {
		healthEvery = time.Second
	}
	tcpfEvery := time.Duration(c.cfg.Performance.TCPFlagRefreshMs) * time.Millisecond
	if tcpfEvery <= 0 {
		tcpfEvery = 5 * time.Second
	}

	if tc.conn == nil {
		flog.Errorf("connection is unexpectedly nil, attempting to recreate connection")
		c, err := tc.createConn()
		if err != nil {
			flog.Errorf("failed to create initial connection: %v", err)
			return nil, fmt.Errorf("failed to create initial connection: %w", err)
		}
		tc.conn = c
		now := time.Now()
		tc.expire = now.Add(300 * time.Second)
		tc.lastHealthCheck = now
		tc.lastTCPFSend = now
	}

	now := time.Now()
	if now.Sub(tc.lastTCPFSend) >= tcpfEvery {
		if err := tc.sendTCPF(tc.conn); err != nil {
			flog.Debugf("failed to refresh TCPF: %v", err)
		} else {
			tc.lastTCPFSend = now
		}
	}

	if forceCheck || now.Sub(tc.lastHealthCheck) >= healthEvery {
		tc.lastHealthCheck = now
		err := tc.conn.Ping(false)
		if err == nil {
			return tc.conn, nil
		}

		flog.Infof("connection lost, recreating transport connection")
		if tc.conn != nil {
			_ = tc.conn.Close()
		}
		c, err := tc.createConn()
		if err != nil {
			flog.Errorf("failed to recreate connection: %v", err)
			return nil, fmt.Errorf("failed to recreate connection: %w", err)
		}
		tc.conn = c
		now = time.Now()
		tc.expire = now.Add(300 * time.Second)
		tc.lastHealthCheck = now
		tc.lastTCPFSend = now
	}
	return tc.conn, nil
}

func (c *Client) newStrm() (tnet.Strm, error) {
	return c.newStrmWithRetry(0)
}

func (c *Client) newStrmWithRetry(attempt int) (tnet.Strm, error) {
	maxAttempts := c.cfg.Performance.MaxRetryAttempts
	if maxAttempts <= 0 {
		maxAttempts = 5
	}

	if attempt >= maxAttempts {
		return nil, fmt.Errorf("failed to create stream after %d attempts", attempt)
	}

	conn, err := c.newConn(attempt > 0)
	if err != nil {
		flog.Debugf("session creation failed (attempt %d/%d), retrying after backoff", attempt+1, maxAttempts)
		backoff := c.calculateRetryBackoff(attempt)
		time.Sleep(backoff)
		return c.newStrmWithRetry(attempt + 1)
	}

	strm, err := conn.OpenStrm()
	if err != nil {
		flog.Debugf("failed to open stream (attempt %d/%d), retrying: %v", attempt+1, maxAttempts, err)
		backoff := c.calculateRetryBackoff(attempt)
		time.Sleep(backoff)
		return c.newStrmWithRetry(attempt + 1)
	}

	return strm, nil
}

func (c *Client) calculateRetryBackoff(attempt int) time.Duration {
	initialBackoff := c.cfg.Performance.RetryInitialBackoffMs
	maxBackoff := c.cfg.Performance.RetryMaxBackoffMs

	if initialBackoff <= 0 {
		initialBackoff = 100
	}
	if maxBackoff <= 0 {
		maxBackoff = 10000
	}

	// Exponential backoff: initialBackoff * 2^attempt
	backoffMs := float64(initialBackoff) * math.Pow(2, float64(attempt))
	if backoffMs > float64(maxBackoff) {
		backoffMs = float64(maxBackoff)
	}

	return time.Duration(backoffMs) * time.Millisecond
}
