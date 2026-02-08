# Optimized Configuration Files

This directory contains example configuration files for paqet.

## Configuration Files

### Standard Examples

- **`server.yaml.example`** - Standard server configuration with performance settings enabled
- **`client.yaml.example`** - Standard client configuration with performance settings enabled

### Optimized for High Performance (1Gbps / 8-core / 8GB RAM)

- **`server.optimized.yaml`** - Server configuration optimized for high throughput
- **`client.optimized.yaml`** - Client configuration optimized for high throughput

## Quick Start

### Using Standard Configuration

1. Copy the example file:
   ```bash
   cp server.yaml.example config.yaml
   ```

2. Edit the configuration and change the marked values:
   - `CHANGE ME:` Network interface, IP addresses, MAC addresses
   - `key:` Secret encryption key

3. Run paqet:
   ```bash
   sudo ./paqet run -c config.yaml
   ```

### Using Optimized Configuration (1Gbps / 8-core / 8GB RAM)

The optimized configurations are pre-tuned for servers with:
- **Network**: 1Gbps bandwidth
- **CPU**: 8 cores
- **RAM**: 8GB

**On Server:**
```bash
cp server.optimized.yaml config.yaml
# Edit CHANGE ME values
sudo ./paqet run -c config.yaml
```

**On Client:**
```bash
cp client.optimized.yaml config.yaml
# Edit CHANGE ME values
sudo ./paqet run -c config.yaml
```

## Key Optimizations in Optimized Configs

The optimized configurations include:

### Performance Settings
- **packet_workers: 8** - Utilizes all CPU cores for packet processing
- **max_concurrent_streams: 15000** (server) / 8000 (client) - Higher concurrency limits
- **enable_connection_pooling: true** (server only) - Reuses TCP connections for 5-10x speed improvement

### Transport Settings
- **conn: 4** - Multiple KCP connections for better parallelism
- **mode: fast2** - Aggressive retransmission for low latency
- **Larger buffers** - 16KB TCP buffers, 8MB SMUX buffers for high throughput

### Network Settings
- **sockbuf: 16777216** - 16MB PCAP buffer for 1Gbps
- **send_queue_size: 2000** - Larger queue for high packet rate

## Performance Expectations

With optimized configuration on 1Gbps / 8-core / 8GB hardware:

- **Throughput**: 500-900 Mbps
- **Latency**: 10-50ms
- **Concurrent Connections**: Up to 15,000 streams
- **CPU Usage**: Distributed across all 8 cores
- **Memory Usage**: 2-4GB under normal load

## Connection Pooling

Connection pooling is **enabled by default** in the optimized server configuration and the standard server example. This is a critical optimization that:

- Reduces connection establishment latency by 5-10x
- Reuses TCP connections to the same targets
- Automatically cleans up idle connections after 90-120 seconds
- Is recommended for production deployments

To **disable** connection pooling (not recommended):
```yaml
performance:
  enable_connection_pooling: false
```

## Customizing Configurations

If your hardware differs from the optimized specs:

### More CPU Cores (16+)
```yaml
performance:
  packet_workers: 16  # Match your core count
transport:
  conn: 8             # More connections
```

### Less RAM (4GB)
```yaml
performance:
  max_concurrent_streams: 5000  # Reduce concurrency
```

### Lower Bandwidth (100Mbps)
```yaml
network:
  pcap:
    sockbuf: 8388608              # 8MB buffer
    send_queue_size: 1000
transport:
  kcp:
    rcvwnd: 1024                  # Smaller windows
    sndwnd: 1024
```

## Troubleshooting

### High Memory Usage
If memory exceeds 6GB:
- Reduce `max_concurrent_streams`
- Reduce `tcp_connection_pool_size`

### Low Throughput
If throughput is below 500Mbps:
- Check CPU usage (should be distributed across cores)
- Increase `send_queue_size`
- Try `mode: "fast3"` for more aggressive retransmission

### Connection Errors
If seeing connection failures:
- Ensure iptables rules are configured (see config comments)
- Check that `key` matches between client and server
- Verify network connectivity with `paqet ping`

## More Information

- **Full Documentation**: See [README.md](../README.md)
- **Performance Guide**: See [docs/PERFORMANCE.md](../docs/PERFORMANCE.md)
- **KCP Modes**: See config comments for detailed explanations

## Important Notes

1. **Security**: Always change the `key` value to a secure random key (use `paqet secret`)
2. **Firewall**: Server requires iptables configuration (see config comments)
3. **Root Access**: paqet requires `sudo` for raw socket access
4. **Testing**: Use `paqet ping` to test connectivity before full deployment
