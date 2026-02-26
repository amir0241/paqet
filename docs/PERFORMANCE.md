# Performance Guide (v1.4)

This document describes the production performance model in `paqet` version `1.4`.

## v1.4 Highlights

- Adaptive defaults now scale by CPU/RAM and by role (`client` vs `server`).
- Connection health checks and TCP-flag refresh intervals are now configurable.
- Retry behavior is bounded and backoff-driven in both stream creation and packet send paths.
- Server-side connection pooling is enabled by default.
- Packet pressure is observable through periodic dropped-packet and queue-depth logs.

## Performance Section

All fields are optional. If omitted, `paqet` applies defaults from `internal/conf/performance.go`.

```yaml
performance:
  max_concurrent_streams: 0
  packet_workers: 0
  stream_worker_pool_size: 0
  enable_connection_pooling: true
  tcp_connection_pool_size: 0
  tcp_connection_idle_timeout: 75
  max_retry_attempts: 6
  retry_initial_backoff_ms: 100
  retry_max_backoff_ms: 5000
  connection_health_check_ms: 1000
  tcp_flag_refresh_ms: 5000
```

## Adaptive Defaults

### `performance.*`

| Field | Server Default | Client Default |
|---|---|---|
| `max_concurrent_streams` | `clamp(cpu*12500, 50000, 100000)` | `clamp(cpu*2500, 10000, 50000)` |
| `packet_workers` | `clamp(GOMAXPROCS, 2, 64)` and minimum `4` | `clamp(GOMAXPROCS, 2, 64)` |
| `stream_worker_pool_size` | `clamp(cpu*2500, 10000, 100000)` | `clamp(cpu*1250, 5000, 50000)` |
| `enable_connection_pooling` | `true` | `false` |
| `tcp_connection_pool_size` | `clamp(cpu*64, 256, 4096)` | `clamp(cpu*16, 64, 512)` |
| `tcp_connection_idle_timeout` | `75s` | `75s` |
| `max_retry_attempts` | `6` | `6` |
| `retry_initial_backoff_ms` | `100` | `100` |
| `retry_max_backoff_ms` | `5000` | `5000` |
| `connection_health_check_ms` | `1000` | `1000` |
| `tcp_flag_refresh_ms` | `5000` | `5000` |

### `network.pcap.*`

Defaults from `internal/conf/pcap.go`:

| Field | Server Default | Client Default |
|---|---|---|
| `sockbuf` | `nextPow2(clamp(ramMB/256, 16, 64)) MB` | `nextPow2(clamp(ramMB/512, 8, 32)) MB` |
| `send_queue_size` | `clamp(cpu*10000, 10000, 100000)` | `clamp(cpu*10000, 10000, 100000)` |
| `max_retries` | `5` | `5` |
| `initial_backoff_ms` | `15` | `15` |
| `max_backoff_ms` | `2000` | `2000` |

### `transport.*`

Defaults from `internal/conf/transport.go`:

| Field | Server Default | Client Default |
|---|---|---|
| `conn` | `1` | QUIC: `clamp(cpu/2,1,4)`, KCP: `clamp(cpu/3,1,3)` |
| `tcpbuf` | `clamp(cpu*16KB, 64KB, 4MB)` | same |
| `udpbuf` | `clamp(cpu*4KB, 16KB, 1MB)` | same |
| `tunbuf` | `clamp(cpu*64KB, 256KB, 16MB)` | same |

## Why It Improves Latency and Stability

- Connection loss is detected quickly through `connection_health_check_ms`.
- PTCPF metadata stays fresh via `tcp_flag_refresh_ms`, reducing stale path behavior.
- Stream creation retries stop at `max_retry_attempts` with exponential backoff.
- Send queue retries use jitter to avoid synchronized retry bursts.
- Packet workers parallelize serialization and TX.
- Connection pooling avoids repeated TCP handshakes to upstream targets.

## Recommended Profiles

### Low-Latency Profile

```yaml
performance:
  max_concurrent_streams: 15000
  packet_workers: 2
  stream_worker_pool_size: 6000
  max_retry_attempts: 5
  retry_initial_backoff_ms: 50
  retry_max_backoff_ms: 1000
  connection_health_check_ms: 400
  tcp_flag_refresh_ms: 2000
  enable_connection_pooling: true
  tcp_connection_pool_size: 256

network:
  pcap:
    send_queue_size: 20000
    max_retries: 4
    initial_backoff_ms: 10
    max_backoff_ms: 500
```

### High-Throughput Profile

```yaml
performance:
  packet_workers: 8
  max_concurrent_streams: 80000
  stream_worker_pool_size: 30000
  enable_connection_pooling: true
  tcp_connection_pool_size: 1024
  tcp_connection_idle_timeout: 120

transport:
  tcpbuf: 262144
  udpbuf: 65536
  tunbuf: 1048576

network:
  pcap:
    sockbuf: 67108864
    send_queue_size: 60000
```

## Observability

`paqet` emits packet pressure warnings every 30 seconds when pressure exists:

- client: `client packet pressure: dropped=..., queue_depth=...`
- server: `server packet pressure: dropped=..., queue_depth=...`

If this appears continuously:

1. Increase `network.pcap.send_queue_size`.
2. Increase `performance.packet_workers`.
3. Increase `network.pcap.sockbuf`.
4. Reduce burst rate or upstream fan-out.

## Migration Notes for v1.4

- Most users do not need manual tuning after upgrade.
- Existing configs remain valid.
- If you previously hardcoded small static buffers/queues, remove overrides and re-test with adaptive defaults first.
- For QUIC-specific tuning, see `docs/QUIC.md`.
