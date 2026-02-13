# KCP Configuration for 1Gbps Bandwidth

This guide explains how to configure paqet with KCP protocol to achieve 1Gbps throughput on high-bandwidth, low-loss networks.

> **⚠️ IMPORTANT WARNING**
>
> The 1Gbps configuration is **highly aggressive** and can **REDUCE your bandwidth** if used incorrectly!
>
> **DO NOT use this configuration** unless:
> - ✅ You have a dedicated 1Gbps+ link
> - ✅ Your network has <50ms latency
> - ✅ Your network has <0.5% packet loss
> - ✅ You have tested and confirmed poor performance with balanced configs
>
> **For most users**, use the **balanced configuration** instead:
> - [`example/client-kcp-balanced.yaml.example`](../example/client-kcp-balanced.yaml.example)
> - [`example/server-kcp-balanced.yaml.example`](../example/server-kcp-balanced.yaml.example)

## Overview

The standard KCP configurations included with paqet are optimized for typical use cases with moderate bandwidth and variable network conditions. The **1Gbps configuration is extremely aggressive** and requires perfect network conditions to work properly. Using it on typical networks can cause packet storms, congestion, and reduced throughput.

## When to Use Which Configuration

| Configuration | Best For | Bandwidth | Latency | Loss |
|--------------|----------|-----------|---------|------|
| **balanced** | **Most users** | 100-500 Mbps | <100ms | <1% |
| standard (client.yaml.example) | Basic setup | 50-200 Mbps | <200ms | <2% |
| **1Gbps** (this guide) | **Dedicated gigabit links** | 500-1500 Mbps | <50ms | <0.5% |

**Start with balanced config** - only move to 1Gbps if you:
1. Verified your network meets the requirements
2. Tested balanced config and it's truly limiting you
3. Have the CPU/system resources to handle the load

## Quick Start

⚠️ **Only proceed if you meet the requirements above!**

Use the provided example configurations:

- **Client**: [`example/client-kcp-1gbps.yaml.example`](../example/client-kcp-1gbps.yaml.example)
- **Server**: [`example/server-kcp-1gbps.yaml.example`](../example/server-kcp-1gbps.yaml.example)

Copy these files to your working directory and customize the network settings (IP addresses, MAC addresses, interface names).

## Configuration Principles

### 1. Understanding Bandwidth-Delay Product (BDP)

The Bandwidth-Delay Product determines how much data can be "in flight" on the network at any given time:

```
BDP = Bandwidth × RTT (Round-Trip Time)
```

**Example**: For 1Gbps with 50ms RTT:
```
BDP = (1000 Mbps × 0.05s) / 8 bits per byte = 6.25 MB
```

This means you need at least 6.25 MB of window space to fully utilize a 1Gbps link with 50ms latency.

### 2. KCP Window Sizing

KCP uses send window (`sndwnd`) and receive window (`rcvwnd`) to control how many packets can be in flight:

```
Window Size (packets) = BDP / MTU
```

**For 1Gbps @ 50ms with MTU 1400**:
```
Required packets = 6.25 MB / 1400 bytes ≈ 4,500 packets
```

The 1Gbps configurations use the maximum KCP window size (32768), which provides ~45 MB of capacity:
```
32768 packets × 1400 bytes = 45.8 MB (enough for 7.3Gbps @ 50ms!)
```

### 3. Key Configuration Parameters

#### KCP Protocol Settings

```yaml
transport:
  kcp:
    mode: "manual"              # Fine-grained control
    nodelay: 1                  # Aggressive retransmission
    interval: 10                # 10ms update interval
    resend: 2                   # Fast retransmit
    nocongestion: 1             # No congestion control
    wdelay: false               # Immediate flush
    acknodelay: true            # Immediate ACK
    mtu: 1400                   # Large packets
    rcvwnd: 32768               # Max receive window
    sndwnd: 32768               # Max send window
```

**Parameter Explanation**:

- **`nodelay: 1`**: Enables aggressive retransmission for lower latency
- **`interval: 10`**: Updates every 10ms (balance of CPU vs responsiveness)
- **`resend: 2`**: Retransmits after 2 duplicate ACKs (fast retransmit)
- **`nocongestion: 1`**: Disables congestion control for maximum speed on dedicated/good links
- **`wdelay: false`**: Sends data immediately without batching (lower latency)
- **`acknodelay: true`**: Sends ACKs immediately (reduces RTT uncertainty)
- **`mtu: 1400`**: Near-maximum packet size (1500 - overhead)
- **`rcvwnd/sndwnd: 32768`**: Maximum window sizes for high BDP

#### Buffer Sizes

```yaml
transport:
  tcpbuf: 262144              # 256KB TCP buffer
  udpbuf: 65536               # 64KB UDP buffer
  kcp:
    smuxbuf: 33554432         # 32MB SMUX buffer
    streambuf: 16777216       # 16MB stream buffer
```

**Why Large Buffers?**
- TCP/UDP buffers reduce system call overhead
- SMUX buffer handles multiplexed stream data
- Stream buffer manages per-connection data

#### Multiple Connections

```yaml
transport:
  conn: 4                     # 4 parallel KCP connections
```

Using multiple connections (load distribution) helps:
- Bypass single-connection bottlenecks
- Better utilize multi-core CPUs
- Improve aggregate throughput

**Recommended values**:
- 1Gbps: 4 connections
- 2-5Gbps: 8 connections
- Lower bandwidth: 1-2 connections

#### Parallel Processing

```yaml
performance:
  packet_workers: 8           # Client: 8 workers
  # packet_workers: 16        # Server: 16 workers
  max_concurrent_streams: 20000   # Client
  # max_concurrent_streams: 50000 # Server
```

More packet workers = better CPU utilization on multi-core systems.

**Guidelines**:
- Client: 4-8 workers (typical 4-8 core systems)
- Server: 8-16 workers (typical 8-16+ core systems)
- Adjust based on `top` CPU monitoring

#### PCAP Buffers

```yaml
network:
  pcap:
    sockbuf: 33554432         # Client: 32MB
    # sockbuf: 67108864       # Server: 64MB
    send_queue_size: 10000    # Client: 10K packets
    # send_queue_size: 15000  # Server: 15K packets
```

Large PCAP buffers prevent packet loss during traffic bursts.

## System Tuning

### Linux Kernel Parameters

KCP runs over UDP, so you need to increase UDP buffer limits:

```bash
# Temporary (until reboot)
sudo sysctl -w net.core.rmem_max=67108864
sudo sysctl -w net.core.wmem_max=67108864
sudo sysctl -w net.core.rmem_default=16777216
sudo sysctl -w net.core.wmem_default=16777216
```

To make permanent, add to `/etc/sysctl.conf`:
```
net.core.rmem_max = 67108864
net.core.wmem_max = 67108864
net.core.rmem_default = 16777216
net.core.wmem_default = 16777216
```

Then apply: `sudo sysctl -p`

### File Descriptor Limits

For servers handling many connections:

```bash
# Check current limit
ulimit -n

# Increase temporarily
ulimit -n 65535
```

To make permanent, edit `/etc/security/limits.conf`:
```
* soft nofile 65535
* hard nofile 65535
```

### Server Firewall Rules

**Critical**: Configure iptables to prevent kernel interference:

```bash
# Replace 9999 with your actual port
sudo iptables -t raw -A PREROUTING -p tcp --dport 9999 -j NOTRACK
sudo iptables -t raw -A OUTPUT -p tcp --sport 9999 -j NOTRACK
sudo iptables -t mangle -A OUTPUT -p tcp --sport 9999 --tcp-flags RST RST -j DROP

# Make persistent
sudo apt install iptables-persistent
sudo netfilter-persistent save
```

## Network Requirements

These configurations are optimized for:

✅ **Good For**:
- High bandwidth: 500Mbps - 2Gbps
- Low latency: <100ms RTT
- Low packet loss: <0.5%
- Stable connections
- Data center / VPS environments
- Good quality ISP connections

❌ **Not Good For**:
- High packet loss networks (>1%)
- Very high latency (>200ms)
- Unstable mobile networks
- Heavily congested networks

For problematic networks:
- Enable FEC (forward error correction)
- Use `fast2` or `fast3` mode instead of manual
- Reduce window sizes
- Consider QUIC instead of KCP

## Performance Expectations

### Theoretical Maximum

With optimal settings and ideal network conditions:
- **Single connection**: 400-600 Mbps (UDP/KCP overhead)
- **4 connections**: 1.2-1.5 Gbps (load distribution)
- **8 connections**: 1.5-2.0 Gbps (with powerful CPU)

### Real-World Performance

Actual performance depends on:
- CPU speed (encryption overhead)
- Network quality (jitter, loss)
- Competing traffic
- OS and hardware

**Typical results** with 1Gbps configurations:
- Good network: 700-900 Mbps aggregate
- Average network: 500-700 Mbps aggregate
- Poor network: 200-400 Mbps aggregate

## Tuning Tips

### For Higher Throughput

1. **Increase connections**: `conn: 6` or `conn: 8`
2. **More packet workers**: Match CPU core count
3. **Larger buffers**: Double tcpbuf/udpbuf/smuxbuf
4. **Faster encryption**: Use `block: "aes-128"` instead of `"aes"`
5. **Disable encryption** (if security allows): `block: "null"`

### For Lower Latency

1. **Reduce interval**: `interval: 5` (more CPU intensive)
2. **Single connection**: `conn: 1`
3. **Smaller buffers**: `tcpbuf: 65536`
4. **Fast mode**: `mode: "fast3"`

### For Lossy Networks

1. **Enable FEC**: Uncomment `dshard: 10` and `pshard: 3`
2. **More aggressive**: `mode: "fast3"` with `resend: 1`
3. **Reduce MTU**: `mtu: 1200` (less loss per packet)

## Monitoring and Testing

### Test Bandwidth

```bash
# Using iperf3 through paqet SOCKS5 proxy
# On server: iperf3 -s
# On client (through proxy):
iperf3 -c SERVER_IP -p 5201 --proxy socks5h://127.0.0.1:1080
```

### Monitor CPU Usage

```bash
# Check paqet CPU usage
top -p $(pgrep paqet)

# If CPU is maxed, increase packet_workers
# If CPU is low, issue is likely network/bandwidth
```

### Monitor Packet Loss

Check paqet logs for dropped packets:
```bash
sudo ./paqet run -c config.yaml 2>&1 | grep -i "drop\|loss"
```

If you see drops, increase `send_queue_size` and `sockbuf`.

## Troubleshooting

### ⚠️ BANDWIDTH IS SLOWER WITH 1GBPS CONFIG

**This is the most common issue!** The 1Gbps config is too aggressive for your network.

**Immediate solution**: Switch to the **balanced configuration**:
```bash
# Use the balanced config instead
cp example/client-kcp-balanced.yaml.example config.yaml
# Edit your network settings and test again
```

**Why this happens**:
1. **Too aggressive retransmission** - Causes packet storms and congestion
2. **No congestion control** - Overwhelms routers/switches
3. **Multiple connections** - Causes packet reordering and drops
4. **Excessive window sizes** - Causes bufferbloat and latency spikes
5. **High CPU overhead** - Can't keep up with packet processing

**Symptoms of over-aggressive configuration**:
- Bandwidth is LOWER than with default config
- High packet loss (>1%)
- Latency spikes or jitter
- Router/switch becomes unresponsive
- CPU usage is very high
- Connection drops or timeouts

**How to fix**:
1. ✅ **Start with balanced config** - Works for 100-500 Mbps
2. If still slow, try standard config (client.yaml.example)
3. Test your network quality:
   ```bash
   # Check latency
   ping -c 100 YOUR_SERVER_IP
   
   # Check packet loss
   ping -c 1000 -i 0.2 YOUR_SERVER_IP | grep loss
   ```
4. Only use 1Gbps config if:
   - Balanced config truly limits you (confirmed with testing)
   - Your network has <50ms latency and <0.5% loss
   - You have a dedicated high-quality link

### Not Reaching 1Gbps

⚠️ **First, are you SURE you should be using the 1Gbps config?**
- If not, switch to balanced config (see above)

If you've confirmed your network meets the requirements:

**Check CPU**: Is paqet using 100% CPU?
- Yes: Add more `packet_workers` or use faster CPU
- No: Issue is elsewhere

**Check encryption**: Encryption is CPU intensive
- Try `block: "aes-128"` for less overhead
- Test with `block: "null"` to rule out encryption bottleneck

**Check window sizes**: Are they at maximum?
- Verify `rcvwnd: 32768` and `sndwnd: 32768`

**Check connections**: Try more parallel connections
- Increase `conn: 4` to `conn: 8`

**Check network**: Is the network actually 1Gbps?
- Test direct speed without paqet
- Check for ISP throttling or QoS

### High Latency

**Check if config is too aggressive**: Try balanced config

**Reduce interval**: Try `interval: 5` or `interval: 20`

**Check network**: KCP adds some latency overhead
- Expect 10-30ms additional latency vs direct connection
- If >50ms extra, check network quality

### Packet Drops

**Check if config is too aggressive**: Try balanced config first!

**Increase PCAP buffers**: Double `sockbuf` and `send_queue_size`

**Check system buffers**: Verify kernel UDP buffer settings

**Check CPU**: High CPU = can't process packets fast enough

**Enable FEC**: If packet loss is consistent (>0.5%), enable FEC:
```yaml
kcp:
  dshard: 10
  pshard: 3
```

## Comparison with QUIC

For 1Gbps bandwidth scenarios:

**Use KCP When**:
- Simple configuration needed
- Single or few connections
- Don't need TLS certificates
- Moderate latency networks (<100ms)

**Use QUIC When**:
- Many concurrent connections (100+)
- Need built-in TLS security
- Very high bandwidth (>2Gbps)
- Production deployments
- See: [`docs/QUIC.md`](QUIC.md) and [`docs/HIGH-LOAD-QUIC.md`](HIGH-LOAD-QUIC.md)

## Security Considerations

### Encryption Performance

AES encryption at 1Gbps requires significant CPU:
- **AES**: ~1-2 CPU cores at 1Gbps
- **AES-128**: ~20% faster than AES
- **null**: No encryption (highest performance)

**Recommendation**: Use `"aes"` or `"aes-128"` for production. Only use `"null"` for testing or trusted networks.

### Key Management

Always use strong keys:
```bash
# Generate secure random key
./paqet secret
```

Change keys regularly and never reuse keys across different deployments.

## Advanced: Manual Fine-Tuning

If the 1Gbps configs don't work well for your specific network, tune manually:

1. **Measure your RTT**:
   ```bash
   ping SERVER_IP
   # Note average RTT
   ```

2. **Calculate required window**:
   ```
   Required_Packets = (Bandwidth_Mbps × RTT_ms) / (MTU × 8)
   Example: (1000 × 50) / (1400 × 8) = 4,464 packets
   ```

3. **Set window sizes**:
   ```yaml
   rcvwnd: 4500  # Round up
   sndwnd: 4500
   ```

4. **Adjust for packet loss**:
   - 0% loss: Use calculated value
   - 0.1-0.5% loss: Add 20% to window size
   - 0.5-1% loss: Add 50% and enable FEC
   - >1% loss: Double window and use fast3 mode + FEC

## Example Deployment

**Scenario**: VPS client → VPS server, 1Gbps links, 30ms RTT, 0.1% loss

**Configuration choices**:
- `mtu: 1400` (standard)
- `rcvwnd/sndwnd: 32768` (max for safety)
- `conn: 4` (load distribution)
- `interval: 10` (balanced)
- `nodelay: 1, resend: 2` (aggressive)
- `nocongestion: 1` (no congestion control)
- `block: "aes"` (secure)
- `packet_workers: 8` (8-core VPS)

**Expected throughput**: 800-1000 Mbps

## References

- [KCP Protocol Documentation](https://github.com/skywind3000/kcp)
- [kcp-go Library](https://github.com/xtaci/kcp-go)
- [paqet Performance Guide](PERFORMANCE.md)
- [paqet QUIC Guide](QUIC.md)

## Summary

To achieve 1Gbps with KCP:

1. ✅ Use provided 1Gbps example configs
2. ✅ Increase system UDP buffer limits
3. ✅ Configure server iptables rules
4. ✅ Use multiple connections (conn: 4)
5. ✅ Enable parallel processing (8-16 workers)
6. ✅ Monitor CPU and adjust workers as needed
7. ✅ Test and tune based on your specific network

For questions or issues, please open an issue on GitHub.
