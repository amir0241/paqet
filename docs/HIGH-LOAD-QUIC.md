# QUIC Under High Connection Pressure (v1.4)

This guide targets deployments with very high concurrency (thousands of active streams/connections).

## Critical Fixes Included in v1.4

1. Timeout-bound QUIC operations:
   - dial timeout: `30s`
   - accept timeout loop: `5s`
   - stream-open timeout: `30s`
   - blocking ping timeout: `10s`
2. Correct context propagation from server listener into accepted QUIC connections.
3. Listener close path now returns real close errors.
4. Accept loop avoids recursive retry behavior and related long-run instability.
5. Graceful shutdown path is more reliable under sustained load.

## Recommended High-Load Baseline

Start from:

- `example/server-quic-high-load.yaml.example`
- `example/client-quic-high-load.yaml.example`

## Server Profile (High Pressure)

```yaml
transport:
  protocol: "quic"
  quic:
    max_idle_timeout: 60
    max_incoming_streams: 50000
    max_incoming_uni_streams: 50000
    initial_stream_receive_window: 10485760
    max_stream_receive_window: 41943040
    initial_connection_receive_window: 31457280
    max_connection_receive_window: 104857600
    keep_alive_period: 15
    enable_0rtt: true

performance:
  max_concurrent_streams: 100000
  packet_workers: 8
  stream_worker_pool_size: 30000
  enable_connection_pooling: true
  tcp_connection_pool_size: 1024
  tcp_connection_idle_timeout: 120
  connection_health_check_ms: 1000
  tcp_flag_refresh_ms: 5000

network:
  pcap:
    sockbuf: 67108864
    send_queue_size: 60000
    max_retries: 5
    initial_backoff_ms: 15
    max_backoff_ms: 2000
```

## Client Profile (High Pressure)

```yaml
transport:
  protocol: "quic"
  conn: 4
  quic:
    max_idle_timeout: 30
    max_incoming_streams: 12000
    max_incoming_uni_streams: 12000
    keep_alive_period: 10
    enable_0rtt: true
    insecure_skip_verify: true

performance:
  max_concurrent_streams: 40000
  packet_workers: 4
  stream_worker_pool_size: 15000
  connection_health_check_ms: 500
  tcp_flag_refresh_ms: 2000

network:
  pcap:
    sockbuf: 33554432
    send_queue_size: 30000
```

## System-Level Requirements

On Linux, apply:

```bash
sudo ulimit -n 1000000
sudo sysctl -w net.core.somaxconn=65535
sudo sysctl -w net.core.netdev_max_backlog=65535
sudo sysctl -w net.core.rmem_max=134217728
sudo sysctl -w net.core.wmem_max=134217728
```

Required `iptables` rules (replace `<PORT>`):

```bash
sudo iptables -t raw -A PREROUTING -p tcp --dport <PORT> -j NOTRACK
sudo iptables -t raw -A OUTPUT -p tcp --sport <PORT> -j NOTRACK
sudo iptables -t mangle -A OUTPUT -p tcp --sport <PORT> --tcp-flags RST RST -j DROP
```

## Operational Signals to Watch

- Repeated `packet pressure` warnings.
- Rising dropped packet counters.
- Queue depth staying above zero.
- Increasing stream-open retries.

If these persist:

1. Increase `network.pcap.send_queue_size`.
2. Increase `performance.packet_workers`.
3. Increase `network.pcap.sockbuf`.
4. Scale out server instances or reduce per-node fan-in.
