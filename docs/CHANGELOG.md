# Changelog

All notable changes are documented in this file.

## [1.4] - 2026-02-26

### Added
- `performance.connection_health_check_ms` for configurable transport health probe cadence.
- `performance.tcp_flag_refresh_ms` for periodic PTCPF metadata refresh on active connections.
- Packet pressure monitoring logs (dropped packets + queue depth) on both client and server.
- High-load QUIC example configs:
  - `example/client-quic-high-load.yaml.example`
  - `example/server-quic-high-load.yaml.example`

### Improved
- Adaptive default tuning by machine capacity and role (`client` vs `server`) for:
  - `performance.max_concurrent_streams`
  - `performance.packet_workers`
  - `performance.stream_worker_pool_size`
  - `performance.tcp_connection_pool_size`
  - `transport.conn`, `transport.tcpbuf`, `transport.udpbuf`, `transport.tunbuf`
  - `network.pcap.sockbuf`, `network.pcap.send_queue_size`
- QUIC defaults tuned for high-pressure operation:
  - larger stream/connection receive windows
  - larger server stream limits
  - role-based keep-alive and idle timeout defaults
  - `enable_0rtt` default enabled
- Connection pooling behavior updated:
  - server default `enable_connection_pooling: true`
  - role-aware default pool size (server larger than client)
- Retry behavior updated:
  - stream creation retries are bounded by `max_retry_attempts`
  - exponential backoff controlled by `retry_initial_backoff_ms` and `retry_max_backoff_ms`
  - send-path retry jitter and bounds through `network.pcap.max_retries`, `initial_backoff_ms`, `max_backoff_ms`

### Fixed
- QUIC dial, accept, and stream-open operations now use explicit timeouts to avoid indefinite blocking.
- QUIC listener/connection context propagation fixed for graceful shutdown and cancellation.
- Listener close path returns underlying close errors instead of silently discarding them.
- Accept loop design avoids recursive timeout retries and related long-run stability risks.
- Send queue retry path handles queue-full and shutdown conditions more safely.

### Notes
- Existing configs continue to work; omitted fields automatically receive updated defaults.
- For migration and tuning details, see:
  - `docs/PERFORMANCE.md`
  - `docs/QUIC.md`
  - `docs/HIGH-LOAD-QUIC.md`
