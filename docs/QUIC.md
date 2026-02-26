# QUIC Transport Guide (v1.4)

This document covers QUIC behavior and tuning in `paqet` version `1.4`.

## v1.4 QUIC Improvements

- Dial is timeout-bound (`30s`) to avoid infinite connect waits.
- Accept loop uses timeout-based polling (`5s`) plus parent-context cancellation.
- Stream open is timeout-bound (`30s`) under load.
- Ping path has timeout control for blocking checks (`10s`).
- Listener context propagates into accepted connections for clean shutdown.
- Listener close now reports underlying close errors.

For high-pressure deployment patterns, also read `docs/HIGH-LOAD-QUIC.md`.

## Defaults

Defaults come from `internal/conf/quic.go` and depend on role.

| Field | Server Default | Client Default |
|---|---|---|
| `max_idle_timeout` | `60s` | `30s` |
| `max_incoming_streams` | `50000` | `5000` |
| `max_incoming_uni_streams` | `50000` | `5000` |
| `initial_stream_receive_window` | `10MB` | `6MB` |
| `max_stream_receive_window` | `40MB` | `24MB` |
| `initial_connection_receive_window` | `30MB` | `15MB` |
| `max_connection_receive_window` | `100MB` | `60MB` |
| `keep_alive_period` | `15s` | `10s` |
| `enable_0rtt` | `true` | `true` |
| `enable_datagrams` | `false` | `false` |

## Minimal QUIC Config

```yaml
transport:
  protocol: "quic"
  quic:
    insecure_skip_verify: true
```

## Low-Latency QUIC Profile

Use this when lower recovery time and faster dead-path detection matter more than maximum memory efficiency.

```yaml
transport:
  protocol: "quic"
  conn: 2
  quic:
    max_idle_timeout: 30
    keep_alive_period: 8
    enable_0rtt: true
    max_incoming_streams: 12000
    max_incoming_uni_streams: 12000

performance:
  connection_health_check_ms: 400
  tcp_flag_refresh_ms: 2000
  max_retry_attempts: 5
  retry_initial_backoff_ms: 50
  retry_max_backoff_ms: 1000
```

## High-Concurrency QUIC Profile

```yaml
transport:
  protocol: "quic"
  conn: 4
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
  packet_workers: 8
  max_concurrent_streams: 80000
  stream_worker_pool_size: 30000
```

## TLS Notes

- Server automatically generates a self-signed certificate at startup.
- Client should set `insecure_skip_verify: true` for self-signed deployments.
- In stricter environments, set `insecure_skip_verify: false` and use `server_name`.

## Tuning Order

1. Keep protocol at `quic`.
2. Start with defaults and run traffic tests.
3. Tune `performance.connection_health_check_ms` and `performance.tcp_flag_refresh_ms` for recovery speed.
4. Tune QUIC flow-control windows for bandwidth.
5. Tune `performance.packet_workers` and `network.pcap.send_queue_size` for burst handling.

## Validation Ranges

- `max_idle_timeout`: `1..600`
- `max_incoming_streams`: `1..100000`
- `max_incoming_uni_streams`: `1..100000`
- `initial_stream_receive_window`: `>= 1MB`
- `max_stream_receive_window`: `>= initial_stream_receive_window`
- `initial_connection_receive_window`: `>= 1MB`
- `max_connection_receive_window`: `>= initial_connection_receive_window`
- `keep_alive_period`: `1..60`

## Troubleshooting

### Handshake Failures

- Check `insecure_skip_verify` and `server_name` settings.
- Verify both sides use `transport.protocol: "quic"`.

### Intermittent Drops or Reconnect Storms

- Lower `connection_health_check_ms` moderately (for example `1000 -> 500`).
- Lower `tcp_flag_refresh_ms` moderately (for example `5000 -> 2000`).
- Increase `network.pcap.send_queue_size`.
- Increase `performance.packet_workers`.

### Throughput Below Expectation

- Increase QUIC receive windows.
- Increase `transport.conn` on the client (up to `4` typically for QUIC).
- Increase `transport.tcpbuf`, `transport.udpbuf`, and `network.pcap.sockbuf`.
