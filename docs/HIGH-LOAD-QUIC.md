# QUIC Transport Under High Connection Pressure - Fixes and Configuration

## Overview

This document describes the bugs fixed and configuration recommendations for running paqet's QUIC transport under high connection pressure (thousands of concurrent connections).

## Critical Bugs Fixed

### 1. Context Management Issues

**Problem**: The QUIC transport was using `context.Background()` for all operations, leading to operations that couldn't be cancelled or timeout-controlled.

**Fixed in**:
- `internal/tnet/quic/dial.go`: Added 30-second timeout to dial operations
- `internal/tnet/quic/listen.go`: Accept now uses context with timeout and proper cancellation
- `internal/tnet/quic/conn.go`: Stream operations now have proper timeouts and context propagation

**Impact**: 
- Prevents indefinite blocking on dial/accept operations
- Enables proper graceful shutdown
- Prevents goroutine leaks during shutdown

### 2. Resource Cleanup Issues

**Problem**: Listener close errors were silently ignored, potentially causing resource leaks.

**Fixed in**:
- `internal/tnet/quic/listen.go`: Close() now properly handles and returns errors from both listener and packet connection

**Impact**:
- Prevents file descriptor leaks
- Ensures proper cleanup under high load

### 3. Goroutine Leak in Listen Loop

**Problem**: Server's listen function spawned a goroutine to close the listener on context cancellation, which could leak if Accept was blocked.

**Fixed in**:
- `internal/server/server.go`: Removed the extra goroutine; listener's Accept now handles context cancellation internally

**Impact**:
- Eliminates potential goroutine leak on shutdown
- Cleaner shutdown behavior

### 4. Stream Operation Timeouts

**Problem**: OpenStrm operations had no timeout, allowing indefinite blocking under high load.

**Fixed in**:
- `internal/tnet/quic/conn.go`: Added 30-second timeout to OpenStrm operations
- Added 10-second timeout to Ping operations

**Impact**:
- Prevents stream exhaustion
- Better detection of connection failures

### 5. Context Propagation from Server to Connections

**Problem**: Connections created their own context rooted at Background(), ignoring server shutdown signals.

**Fixed in**:
- `internal/tnet/quic/conn.go`: Added `newConnWithContext()` to create connections that inherit parent context
- `internal/tnet/quic/listen.go`: Listener now passes its context to new connections
- `internal/server/server.go`: Server sets context on QUIC listener

**Impact**:
- Stream operations now respect server shutdown
- Proper cancellation cascade during shutdown

## Configuration for High Connection Pressure

### Server Configuration

Use the new example file: `example/server-quic-high-load.yaml.example`

Key settings:
```yaml
transport:
  quic:
    max_idle_timeout: 60                      # Increased from 30s
    max_incoming_streams: 50000               # Increased from 10000
    max_incoming_uni_streams: 50000           # Increased from 10000
    initial_stream_receive_window: 10485760   # 10 MB (increased from 6 MB)
    max_stream_receive_window: 41943040       # 40 MB (increased from 24 MB)
    initial_connection_receive_window: 31457280 # 30 MB (increased from 15 MB)
    max_connection_receive_window: 104857600  # 100 MB (increased from 60 MB)
    keep_alive_period: 15                     # Increased from 10s

performance:
  max_concurrent_streams: 50000               # Increased from 10000
  packet_workers: 8                           # Increased from default
  stream_worker_pool_size: 5000               # Increased from 1000
  enable_connection_pooling: true             # Recommended for high load
  tcp_connection_pool_size: 500               # Increased from 100
```

### Client Configuration

Use the new example file: `example/client-quic-high-load.yaml.example`

Key settings:
```yaml
transport:
  quic:
    max_idle_timeout: 60                      # Increased from 30s
    max_incoming_streams: 10000               # Increased from 5000
    max_incoming_uni_streams: 10000           # Increased from 5000
    keep_alive_period: 15                     # Increased from 10s

performance:
  max_concurrent_streams: 10000               # Increased from 5000
  stream_worker_pool_size: 2000               # Increased from 1000
```

## System-Level Tuning

For production deployments under high load, apply these system-level optimizations:

### 1. File Descriptor Limits

```bash
# Increase file descriptor limit
sudo ulimit -n 1000000

# Or permanently in /etc/security/limits.conf:
* soft nofile 1000000
* hard nofile 1000000
```

### 2. Network Stack Tuning

```bash
# Increase connection queue sizes
sudo sysctl -w net.core.somaxconn=65535
sudo sysctl -w net.ipv4.tcp_max_syn_backlog=65535
sudo sysctl -w net.core.netdev_max_backlog=65535

# Optimize memory and buffer sizes
sudo sysctl -w net.core.rmem_max=134217728
sudo sysctl -w net.core.wmem_max=134217728
sudo sysctl -w net.ipv4.tcp_rmem='4096 87380 134217728'
sudo sysctl -w net.ipv4.tcp_wmem='4096 65536 134217728'

# Enable TCP fast open
sudo sysctl -w net.ipv4.tcp_fastopen=3

# Reduce TIME_WAIT connections
sudo sysctl -w net.ipv4.tcp_fin_timeout=15
sudo sysctl -w net.ipv4.tcp_tw_reuse=1
```

To make these changes persistent, add them to `/etc/sysctl.conf`.

### 3. Required iptables Rules

**CRITICAL**: These rules are required for the server to function properly:

```bash
# Replace <PORT> with your server listen port (e.g., 9999)

# 1. Bypass connection tracking
sudo iptables -t raw -A PREROUTING -p tcp --dport <PORT> -j NOTRACK
sudo iptables -t raw -A OUTPUT -p tcp --sport <PORT> -j NOTRACK

# 2. Prevent kernel RST packets
sudo iptables -t mangle -A OUTPUT -p tcp --sport <PORT> --tcp-flags RST RST -j DROP
```

## Performance Tuning Guidelines

### Memory vs Concurrency Trade-offs

1. **High Bandwidth, Moderate Connections**: Increase window sizes
   - `initial_stream_receive_window`: 10-20 MB
   - `max_stream_receive_window`: 40-80 MB

2. **Many Connections, Moderate Bandwidth**: Increase stream limits, moderate windows
   - `max_incoming_streams`: 50000+
   - Keep default window sizes

3. **Both High Bandwidth and Many Connections**: Increase all settings
   - Use the high-load example configurations
   - Monitor memory usage and adjust accordingly

### CPU Utilization

- `packet_workers`: Set to number of CPU cores, minimum 4 for servers
- `stream_worker_pool_size`: Increase for better parallelism (1000-5000 for servers)
- `max_concurrent_streams`: Should be higher than worker pool size

### Connection Pooling

Enable `enable_connection_pooling: true` on servers to reuse TCP connections to upstream targets:
- Reduces connection establishment overhead
- Improves latency for repeated connections
- Set `tcp_connection_pool_size` to 500-1000 for high-pressure servers

## Monitoring and Troubleshooting

### Signs of Resource Exhaustion

1. **"too many open files"**: Increase file descriptor limits
2. **Increasing latency**: Increase worker pools or concurrent stream limits
3. **Connection timeouts**: Check network tuning, increase idle timeouts
4. **Memory pressure**: Reduce window sizes or stream limits

### Recommended Monitoring

Monitor these metrics:
- Open file descriptors: `lsof -p <PID> | wc -l`
- Active connections: `netstat -an | grep :9999 | wc -l`
- Memory usage: `ps aux | grep paqet`
- CPU usage: `top -p <PID>`

### Logging

For production under high load:
```yaml
log:
  level: "warn"  # or "error" to reduce overhead
```

## Testing Under Load

Use tools like `ab`, `wrk`, or custom scripts to simulate high connection pressure:

```bash
# Example with wrk (HTTP load testing)
wrk -t 12 -c 1000 -d 30s --latency http://target-via-proxy

# Monitor paqet server during test
watch -n 1 'ps aux | grep paqet'
```

## Migration from KCP to QUIC

If migrating from KCP under high load:

1. QUIC handles many connections better than KCP
2. QUIC has lower CPU overhead for multiplexing
3. QUIC provides better congestion control
4. Start with the high-load example configurations
5. Gradually tune based on your specific load patterns

## Summary

The fixes address critical bugs in:
- Context management (preventing goroutine leaks)
- Timeout handling (preventing indefinite blocking)
- Resource cleanup (preventing file descriptor leaks)
- Graceful shutdown (proper cancellation propagation)

Combined with the optimized configuration and system tuning, paqet can now handle thousands of concurrent connections reliably under high pressure.
