# TUN Mode Architecture

## Overview

TUN mode in paqet creates a **secure layer 3 VPN tunnel** by establishing a private network overlay between client and server. **All TUN traffic flows through paqet's encrypted transport layer (KCP or QUIC)**, ensuring that packets are protected by the same raw packet encapsulation and encryption used for SOCKS5 mode.

## Common Misconception

**Question:** "Does TUN mode use paqet to create the private network?"

**Answer:** **YES, absolutely.** TUN mode fully utilizes paqet's encrypted transport:
- All packets read from the TUN device are sent through paqet streams
- These streams use KCP or QUIC transport with encryption
- Packets are encapsulated in raw TCP packets just like SOCKS5 traffic
- The server decrypts and forwards packets to its TUN device
- Return traffic follows the same encrypted path in reverse

TUN mode does NOT create a direct network connection between client and server that bypasses paqet. The TUN interfaces are purely virtual and serve as entry/exit points for the encrypted tunnel.

## Architecture Diagram

```
┌────────────────── CLIENT ──────────────────┐
│                                             │
│  Application Layer                          │
│       │ (sends IP packet to 10.0.8.2)      │
│       ↓                                     │
│  ┌─────────────────────────────┐           │
│  │  TUN Device (tun0)          │           │
│  │  IP: 10.0.8.1/24            │           │
│  │  (Virtual Network Interface)│           │
│  └─────────────────────────────┘           │
│       │ Read IP packet                      │
│       ↓                                     │
│  ┌─────────────────────────────┐           │
│  │  tunnel.Handler.Start()     │           │
│  │  - Reads from TUN device    │           │
│  │  - Writes to paqet stream   │           │
│  └─────────────────────────────┘           │
│       │ buffer.CopyTUN()                    │
│       ↓                                     │
│  ┌─────────────────────────────┐           │
│  │  client.TUN()               │           │
│  │  - Creates paqet stream     │           │
│  │  - Sends PTUN header        │           │
│  └─────────────────────────────┘           │
│       │ Encrypted Stream                    │
│       ↓                                     │
│  ┌─────────────────────────────┐           │
│  │  KCP/QUIC Transport Layer   │           │
│  │  - AES encryption           │           │
│  │  - Reliable delivery        │           │
│  │  - Connection multiplexing  │           │
│  └─────────────────────────────┘           │
│       │ Encrypted data                      │
│       ↓                                     │
│  ┌─────────────────────────────┐           │
│  │  Raw Packet Socket          │           │
│  │  - Crafted TCP packets      │           │
│  │  - pcap injection           │           │
│  └─────────────────────────────┘           │
│       │                                     │
└───────┼─────────────────────────────────────┘
        │
        │  Internet (Raw TCP Packets)
        │  Encrypted KCP/QUIC payload
        │
┌───────┼───────────── SERVER ────────────────┐
│       │                                     │
│       ↓                                     │
│  ┌─────────────────────────────┐           │
│  │  Raw Packet Socket          │           │
│  │  - pcap capture             │           │
│  │  - TCP packet parsing       │           │
│  └─────────────────────────────┘           │
│       │ Encrypted data                      │
│       ↓                                     │
│  ┌─────────────────────────────┐           │
│  │  KCP/QUIC Transport Layer   │           │
│  │  - AES decryption           │           │
│  │  - Stream demultiplexing    │           │
│  └─────────────────────────────┘           │
│       │ Decrypted Stream                    │
│       ↓                                     │
│  ┌─────────────────────────────┐           │
│  │  server.handleTUNProtocol() │           │
│  │  - Reads PTUN header        │           │
│  │  - Bidirectional relay      │           │
│  └─────────────────────────────┘           │
│       │ buffer.CopyTUN()                    │
│       ↓                                     │
│  ┌─────────────────────────────┐           │
│  │  TUN Device (tun0)          │           │
│  │  IP: 10.0.8.2/24            │           │
│  │  (Virtual Network Interface)│           │
│  └─────────────────────────────┘           │
│       │ Write IP packet                     │
│       ↓                                     │
│  OS Network Stack                           │
│  (packet reaches 10.0.8.2)                  │
│                                             │
└─────────────────────────────────────────────┘
```

## Detailed Packet Flow

### Outbound (Client → Server)

1. **Application sends packet** to `10.0.8.2` (server's TUN IP)
2. **OS routing** directs packet to `tun0` interface (client TUN device)
3. **TUN device** captures the IP packet
4. **tunnel.Handler** reads packet from TUN device
5. **buffer.CopyTUN()** copies packet to paqet stream
6. **client.TUN()** stream with `PTUN` protocol header
7. **KCP/QUIC transport** encrypts packet (AES encryption)
8. **Raw packet socket** encapsulates in crafted TCP packet
9. **pcap** injects packet onto network
10. **Internet** transmits encrypted raw TCP packet
11. **Server's pcap** captures incoming TCP packet
12. **KCP/QUIC transport** decrypts payload
13. **server.handleTUNProtocol()** receives decrypted packet
14. **buffer.CopyTUN()** copies to server's TUN device
15. **Server TUN device** writes packet to OS network stack
16. **Packet arrives** at destination (10.0.8.2)

### Inbound (Server → Client)

The process is reversed for return traffic:
- Server TUN device → paqet stream → encrypted transport → client TUN device

## Key Components

### 1. TUN Device (`internal/tunnel/tun.go`)

- Creates virtual network interface using `water` library
- Configured with IP address (e.g., `10.0.8.1/24`)
- Provides Read/Write methods for IP packets
- Platform-specific configuration (Linux: `ip` commands, macOS: `ifconfig`)

**Important:** The TUN device is purely local. It does NOT create any direct network connection. All packets are handed off to the tunnel handler.

### 2. Tunnel Handler (`internal/tunnel/handler.go`)

- Manages the TUN tunnel lifecycle
- Creates paqet stream via `client.TUN()`
- Sets up bidirectional relay: TUN ↔ paqet stream
- Uses `buffer.CopyTUN()` with 256KB buffer pool for efficiency
- Handles context cancellation for clean shutdown

### 3. Client TUN Method (`internal/client/tun.go`)

- Creates new paqet stream (encrypted KCP/QUIC connection)
- Sends `PTUN` protocol header to identify stream type
- Returns stream for packet relay
- This is where TUN traffic enters paqet's encrypted transport

### 4. Server TUN Handler (`internal/server/tun.go`)

- Receives `PTUN` protocol header from stream
- Validates TUN is enabled on server
- Sets up bidirectional relay: paqet stream ↔ server TUN device
- Decrypted packets are written to server's TUN device

## Security Model

### Encryption

TUN traffic is encrypted by the configured transport protocol:

**KCP Transport:**
- Block cipher (AES, etc.) configured via `transport.kcp.block`
- Secret key authentication via `transport.kcp.key`
- All packets encrypted before transmission

**QUIC Transport:**
- TLS 1.3 encryption
- Certificate-based authentication
- 0-RTT support for established connections

### Firewall Bypass

Like SOCKS5 mode, TUN mode uses raw packet injection via pcap:
- Bypasses host OS TCP/IP stack
- Avoids kernel connection tracking
- Cannot be blocked by standard firewall rules
- Requires iptables rules on server to prevent RST packets

## Comparison with SOCKS5 Mode

| Feature | SOCKS5 Mode | TUN Mode |
|---------|-------------|----------|
| **Transport** | paqet (KCP/QUIC) | paqet (KCP/QUIC) |
| **Encryption** | Yes | Yes |
| **Layer** | Application (Layer 7) | Network (Layer 3) |
| **Protocols** | TCP/UDP via SOCKS5 | Any IP protocol* (TCP, UDP, ICMP, etc.) |
| **Configuration** | Per-application proxy | System-wide routing |
| **Transparency** | Requires SOCKS5 support | Transparent to applications |
| **Raw Packets** | Yes | Yes |

*Note: TUN mode forwards raw IP packets at layer 3. While it technically supports any IP protocol, practical support depends on the server's network configuration and routing capabilities. Common protocols (TCP, UDP, ICMP) work without issues.

**Both modes use paqet's encrypted transport.** The difference is how applications access the tunnel:
- SOCKS5: Applications connect to local proxy
- TUN: Applications use routing table to send packets through TUN interface

## Use Cases

### 1. Point-to-Point VPN
```yaml
# Client: 10.0.8.1
# Server: 10.0.8.2
# Applications can directly communicate via these IPs
```

### 2. Remote Network Access
```bash
# Route entire subnet through tunnel
ip route add 192.168.1.0/24 via 10.0.8.2 dev tun0
```

### 3. Secure Service Access
```bash
# Access server's services through tunnel
curl http://10.0.8.2:8080
ssh user@10.0.8.2
```

### 4. Network Segmentation
- Create isolated network overlay
- Connect remote segments securely
- Build mesh networks

## Configuration Example

### Client (`client-tun.yaml`)
```yaml
role: "client"

tun:
  enabled: true
  name: "tun0"
  addr: "10.0.8.1/24"
  mtu: 1400

network:
  interface: "en0"
  ipv4:
    addr: "192.168.1.100:0"
    router_mac: "aa:bb:cc:dd:ee:ff"

server:
  addr: "203.0.113.10:9999"

transport:
  protocol: "kcp"
  kcp:
    block: "aes"
    key: "your-secret-key"
```

### Server (`server-tun.yaml`)
```yaml
role: "server"

listen:
  addr: ":9999"

tun:
  enabled: true
  name: "tun0"
  addr: "10.0.8.2/24"
  mtu: 1400

network:
  interface: "eth0"
  ipv4:
    addr: "203.0.113.10:9999"
    router_mac: "aa:bb:cc:dd:ee:ff"

transport:
  protocol: "kcp"
  kcp:
    block: "aes"
    key: "your-secret-key"
```

## Performance Considerations

### MTU Size
- System default: 1500 bytes (used by paqet when not specified)
- Recommended: 1400 bytes or lower (accounts for paqet transport overhead)
- Lower values may be needed for heavily encapsulated networks or high packet loss scenarios
- Set via `tun.mtu` in configuration file

### Buffer Size
- Uses 256KB buffer pool via `buffer.CopyTUN()`
- Optimized for high throughput
- Avoids frequent small allocations

### Transport Selection
- **KCP:** Better for high-loss networks, lower latency
- **QUIC:** Better for high bandwidth, many connections

## Troubleshooting

### TUN Traffic Not Working

1. **Verify TUN creation:**
   ```bash
   ip addr show tun0  # Linux
   ifconfig tun0      # macOS
   ```

2. **Check routing:**
   ```bash
   ip route | grep tun0  # Linux
   netstat -rn | grep tun0  # macOS
   ```

3. **Test connectivity:**
   ```bash
   ping 10.0.8.2  # From client to server
   ping 10.0.8.1  # From server to client
   ```

4. **Enable debug logging:**
   ```yaml
   log:
     level: "debug"
   ```

5. **Verify paqet connection:**
   - TUN mode requires successful paqet connection
   - Check that transport (KCP/QUIC) is working
   - Verify iptables rules on server

### Common Issues

**"TUN stream received but TUN is not enabled on server"**
- Ensure `tun.enabled: true` on both client and server
- Verify TUN configuration is present

**"Failed to create TUN device"**
- Run with `sudo` (root privileges required)
- Check platform support (Linux/macOS)
- Verify no conflicting TUN device exists

**"Packets not reaching destination"**
- Verify IP addresses are in same subnet
- Check MTU settings
- Enable IP forwarding if routing between networks

## Conclusion

TUN mode in paqet **absolutely uses paqet's encrypted transport** to create a secure private network. The TUN devices are virtual interfaces that serve as entry and exit points for the encrypted tunnel. All traffic flows through paqet's KCP or QUIC transport with full encryption, just like SOCKS5 mode.

The architecture ensures that:
- ✅ All TUN packets are encrypted by paqet's transport layer
- ✅ Traffic bypasses host TCP/IP stack via raw packets
- ✅ Secure layer 3 VPN tunnel is established
- ✅ Private network overlay is created between client and server

For more information, see:
- [README.md](../README.md) - TUN mode usage guide
- [example/client-tun.yaml.example](../example/client-tun.yaml.example) - Client configuration
- [example/server-tun.yaml.example](../example/server-tun.yaml.example) - Server configuration
