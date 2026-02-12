# Python Migration Guide for paqet

This document outlines the complete strategy and roadmap for creating a Python version of the paqet project.

## Overview

paqet is a Go-based bidirectional packet-level proxy that uses raw sockets for packet capture and injection, with support for KCP and QUIC transport protocols. This guide provides a comprehensive roadmap for porting the entire project to Python.

## Project Statistics

- **Language**: Go 1.25+
- **Files**: 77 Go source files
- **Lines of Code**: ~5,685 lines
- **Test Files**: 6 test files
- **Key Dependencies**: pcap, gopacket, kcp-go, quic-go, smux, cobra

## Architecture Overview

### Current Go Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      CLI Layer (cobra)                   │
│  Commands: run, ping, dump, secret, version, iface      │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│                   Configuration Layer                    │
│    YAML parsing, validation, network settings           │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│                    Application Layer                     │
│         Client Mode          │      Server Mode          │
│  - SOCKS5 Proxy             │  - Listen for connections │
│  - Port Forwarding          │  - Forward to targets     │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│                    Transport Layer                       │
│         KCP Transport        │      QUIC Transport       │
│  - Encryption (AES/SM4)     │  - TLS 1.3                │
│  - Multiplexing (smux)      │  - 0-RTT support          │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│                   Raw Socket Layer                       │
│  - pcap packet capture      │  - Packet injection       │
│  - TCP packet crafting      │  - Checksum calculation   │
└─────────────────────────────────────────────────────────┘
```

## Python Technology Stack

### Core Dependencies Mapping

| Go Library | Python Equivalent | Purpose | Status |
|------------|------------------|---------|--------|
| `gopacket/gopacket` | `scapy` | Packet crafting/parsing | ✅ Available |
| `libpcap` | `pypcap` or `pcapy` | Raw packet capture | ✅ Available |
| `kcp-go` | Custom implementation | KCP protocol | ⚠️ Need to implement |
| `quic-go` | `aioquic` | QUIC protocol | ✅ Available |
| `smux` | Custom implementation | Stream multiplexing | ⚠️ Need to implement |
| `cobra` | `click` or `argparse` | CLI framework | ✅ Available |
| `go-yaml` | `PyYAML` | YAML parsing | ✅ Available |
| `socks5` | `PySocks` or custom | SOCKS5 server | ✅ Available |
| `crypto` | `cryptography` | Encryption | ✅ Available |

### Recommended Python Stack

```python
# Core dependencies
scapy>=2.5.0          # Packet crafting and parsing
pypcap>=1.2.3         # Raw packet capture (alternative: pcapy-ng)
aioquic>=0.9.21       # QUIC protocol implementation
click>=8.1.0          # CLI framework
PyYAML>=6.0           # YAML configuration
cryptography>=41.0.0  # Encryption and crypto operations
asyncio               # Async I/O (built-in)

# Network and protocol
aiohttp>=3.8.0        # Async HTTP client/server
PySocks>=1.7.1        # SOCKS proxy support

# Utilities
dataclasses           # Configuration objects (built-in Python 3.7+)
typing                # Type hints (built-in)
logging               # Logging (built-in)
```

## Module-by-Module Porting Guide

### 1. Configuration Module (`internal/conf`)

**Go Files**: `conf.go`, `kcp.go`, `tcp.go`, `socks.go`, `performance.go`, `validation.go`

**Python Equivalent**:
```python
# paqet/config/
#   __init__.py
#   config.py         # Main configuration classes
#   kcp.py           # KCP-specific config
#   quic.py          # QUIC-specific config
#   network.py       # Network settings
#   validation.py    # Config validation

from dataclasses import dataclass
from typing import Optional, List
import yaml

@dataclass
class NetworkConfig:
    interface: str
    guid: Optional[str] = None
    ipv4: Optional['IPv4Config'] = None

@dataclass
class KCPConfig:
    block: str = "aes"
    key: str = ""
    data_shard: int = 10
    parity_shard: int = 3

@dataclass
class Config:
    role: str  # "client" or "server"
    log: 'LogConfig'
    network: NetworkConfig
    transport: 'TransportConfig'
    # ... more fields
```

### 2. Packet Handling (`internal/socket`, `internal/pkg/buffer`)

**Go Files**: `socket.go`, `handle.go`, `send_handle.go`, `recv_handle.go`

**Python Equivalent using scapy**:
```python
# paqet/socket/
#   __init__.py
#   socket.py        # Raw socket management
#   handler.py       # Packet send/receive
#   buffer.py        # Buffer management

from scapy.all import *
import pypcap

class RawSocket:
    def __init__(self, interface: str):
        self.interface = interface
        self.pcap = pypcap.pcap(interface, snaplen=65535, promisc=True)
    
    def send_packet(self, packet: bytes):
        """Inject raw packet using pcap"""
        self.pcap.inject(packet)
    
    def recv_packet(self, timeout: float = 1.0) -> Optional[bytes]:
        """Capture packet from interface"""
        self.pcap.setnonblock(True)
        for ts, pkt in self.pcap.readpkts():
            return pkt
        return None

class TCPPacketHandler:
    def craft_tcp_packet(self, src_ip, dst_ip, src_port, dst_port, 
                        payload, flags="PA", seq=0, ack=0):
        """Craft TCP packet with custom flags"""
        ip = IP(src=src_ip, dst=dst_ip)
        tcp = TCP(sport=src_port, dport=dst_port, 
                 flags=flags, seq=seq, ack=ack)
        return ip/tcp/Raw(load=payload)
```

### 3. KCP Protocol Implementation

**Go Files**: Uses `kcp-go` library

**Python Implementation Strategy**:

Since there's no mature KCP library for Python, you'll need to implement it or use a minimal wrapper:

```python
# paqet/transport/kcp.py

class KCPConnection:
    """
    Minimal KCP protocol implementation
    Based on the KCP specification: https://github.com/skywind3000/kcp
    """
    
    def __init__(self, conv_id: int, output_callback):
        self.conv = conv_id
        self.snd_una = 0  # Unacknowledged send sequence
        self.snd_nxt = 0  # Next send sequence
        self.rcv_nxt = 0  # Next expected receive sequence
        self.output = output_callback
        # ... more KCP state
    
    def send(self, data: bytes) -> int:
        """Send data through KCP"""
        # Implement KCP send logic
        pass
    
    def recv(self) -> Optional[bytes]:
        """Receive data from KCP"""
        # Implement KCP receive logic
        pass
    
    def input(self, data: bytes):
        """Feed received packet into KCP"""
        # Implement KCP input processing
        pass
    
    def update(self, current: int):
        """Update KCP state machine"""
        # Implement KCP update logic
        pass
```

**Alternative**: Use a C-based KCP library with Python bindings (e.g., ctypes or Cython).

### 4. QUIC Transport (`internal/tnet/quic`)

**Python using aioquic**:
```python
# paqet/transport/quic.py

from aioquic.asyncio import connect, serve
from aioquic.quic.configuration import QuicConfiguration

class QUICTransport:
    def __init__(self, config: 'QuicConfig'):
        self.config = self._build_quic_config(config)
    
    def _build_quic_config(self, config):
        quic_config = QuicConfiguration(is_client=True)
        quic_config.load_cert_chain(config.cert, config.key)
        return quic_config
    
    async def connect(self, host: str, port: int):
        """Connect to QUIC server"""
        async with connect(host, port, configuration=self.config) as client:
            # Handle connection
            pass
    
    async def serve(self, host: str, port: int):
        """Start QUIC server"""
        await serve(host, port, configuration=self.config,
                   create_protocol=self._create_protocol)
```

### 5. SOCKS5 Proxy (`internal/socks`)

**Python Implementation**:
```python
# paqet/socks5/
#   __init__.py
#   server.py        # SOCKS5 server
#   client.py        # SOCKS5 client handling

import asyncio
import struct

class SOCKS5Server:
    def __init__(self, host: str, port: int, transport_handler):
        self.host = host
        self.port = port
        self.transport = transport_handler
    
    async def handle_client(self, reader, writer):
        """Handle SOCKS5 client connection"""
        # 1. Handle SOCKS5 handshake
        version = await reader.read(1)
        if version != b'\x05':
            writer.close()
            return
        
        # 2. No authentication method
        nmethods = await reader.read(1)
        methods = await reader.read(ord(nmethods))
        writer.write(b'\x05\x00')  # No auth required
        await writer.drain()
        
        # 3. Handle CONNECT request
        # ... implement SOCKS5 protocol
        
    async def start(self):
        """Start SOCKS5 server"""
        server = await asyncio.start_server(
            self.handle_client, self.host, self.port)
        async with server:
            await server.serve_forever()
```

### 6. Client Mode (`internal/client`)

**Python Structure**:
```python
# paqet/client/
#   __init__.py
#   client.py        # Main client logic
#   proxy.py         # Proxy handling
#   forward.py       # Port forwarding

class PaqetClient:
    def __init__(self, config: Config):
        self.config = config
        self.raw_socket = RawSocket(config.network.interface)
        self.transport = self._init_transport()
        self.socks5 = None
        self.forwarders = []
    
    def _init_transport(self):
        if self.config.transport.protocol == "kcp":
            return KCPTransport(self.config.transport.kcp)
        elif self.config.transport.protocol == "quic":
            return QUICTransport(self.config.transport.quic)
    
    async def start(self):
        """Start client proxy"""
        # 1. Initialize raw socket listener
        asyncio.create_task(self._handle_raw_packets())
        
        # 2. Start SOCKS5 proxy if configured
        if self.config.socks5:
            for socks_config in self.config.socks5:
                self.socks5 = SOCKS5Server(
                    socks_config.listen.split(':')[0],
                    int(socks_config.listen.split(':')[1]),
                    self.transport
                )
                asyncio.create_task(self.socks5.start())
        
        # 3. Start port forwarders if configured
        # ... setup forwarders
```

### 7. Server Mode (`internal/server`)

**Python Structure**:
```python
# paqet/server/
#   __init__.py
#   server.py        # Main server logic

class PaqetServer:
    def __init__(self, config: Config):
        self.config = config
        self.raw_socket = RawSocket(config.network.interface)
        self.transport = self._init_transport()
        self.connections = {}
    
    async def start(self):
        """Start server"""
        # 1. Listen for raw packets
        asyncio.create_task(self._handle_raw_packets())
        
        # 2. Handle transport layer connections
        if self.config.transport.protocol == "kcp":
            await self._start_kcp_server()
        elif self.config.transport.protocol == "quic":
            await self._start_quic_server()
    
    async def _handle_connection(self, conn):
        """Handle incoming connection and forward to target"""
        # Read destination from transport
        # Create connection to target
        # Relay data bidirectionally
        pass
```

### 8. CLI Commands (`cmd/`)

**Python with Click**:
```python
# paqet/cli/
#   __init__.py
#   main.py          # Main CLI entry point
#   run.py           # run command
#   ping.py          # ping command
#   dump.py          # dump command
#   secret.py        # secret generation
#   version.py       # version info

import click
from paqet.client import PaqetClient
from paqet.server import PaqetServer

@click.group()
def cli():
    """paqet - Packet-level proxy with KCP/QUIC transport"""
    pass

@cli.command()
@click.option('-c', '--config', required=True, help='Configuration file')
def run(config):
    """Start paqet client or server"""
    cfg = load_config(config)
    
    if cfg.role == "client":
        client = PaqetClient(cfg)
        asyncio.run(client.start())
    elif cfg.role == "server":
        server = PaqetServer(cfg)
        asyncio.run(server.start())

@cli.command()
def secret():
    """Generate a secret key"""
    import secrets
    key = secrets.token_hex(32)
    click.echo(f"Generated secret key: {key}")

@cli.command()
@click.option('-c', '--config', required=True, help='Configuration file')
def ping(config):
    """Send test packet to server"""
    # Implement ping functionality
    pass

@cli.command()
@click.option('-p', '--port', required=True, type=int, help='Port to monitor')
@click.option('-i', '--interface', default='eth0', help='Network interface')
def dump(port, interface):
    """Dump and decode packets"""
    # Implement packet dumping with scapy
    from scapy.all import sniff
    
    def packet_handler(pkt):
        if TCP in pkt and (pkt[TCP].sport == port or pkt[TCP].dport == port):
            pkt.show()
    
    sniff(iface=interface, prn=packet_handler, filter=f"tcp port {port}")

if __name__ == '__main__':
    cli()
```

### 9. Logging (`internal/flog`)

**Python Logging**:
```python
# paqet/logging/
#   __init__.py
#   logger.py

import logging
import sys

def setup_logger(level: str = "info") -> logging.Logger:
    """Setup application logger"""
    level_map = {
        "none": logging.CRITICAL + 1,
        "debug": logging.DEBUG,
        "info": logging.INFO,
        "warn": logging.WARNING,
        "error": logging.ERROR,
        "fatal": logging.CRITICAL,
    }
    
    logger = logging.getLogger("paqet")
    logger.setLevel(level_map.get(level.lower(), logging.INFO))
    
    handler = logging.StreamHandler(sys.stdout)
    formatter = logging.Formatter(
        '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    handler.setFormatter(formatter)
    logger.addHandler(handler)
    
    return logger
```

## Project Structure

```
paqet-python/
├── paqet/
│   ├── __init__.py
│   ├── cli/
│   │   ├── __init__.py
│   │   ├── main.py
│   │   ├── run.py
│   │   ├── ping.py
│   │   ├── dump.py
│   │   ├── secret.py
│   │   └── version.py
│   ├── config/
│   │   ├── __init__.py
│   │   ├── config.py
│   │   ├── kcp.py
│   │   ├── quic.py
│   │   ├── network.py
│   │   └── validation.py
│   ├── client/
│   │   ├── __init__.py
│   │   ├── client.py
│   │   ├── proxy.py
│   │   └── forward.py
│   ├── server/
│   │   ├── __init__.py
│   │   └── server.py
│   ├── transport/
│   │   ├── __init__.py
│   │   ├── kcp.py
│   │   ├── quic.py
│   │   └── multiplexer.py
│   ├── socket/
│   │   ├── __init__.py
│   │   ├── socket.py
│   │   ├── handler.py
│   │   └── buffer.py
│   ├── socks5/
│   │   ├── __init__.py
│   │   ├── server.py
│   │   └── client.py
│   ├── protocol/
│   │   ├── __init__.py
│   │   └── protocol.py
│   ├── crypto/
│   │   ├── __init__.py
│   │   └── encryption.py
│   └── utils/
│       ├── __init__.py
│       ├── hash.py
│       ├── pool.py
│       └── errors.py
├── tests/
│   ├── __init__.py
│   ├── test_config.py
│   ├── test_socket.py
│   ├── test_kcp.py
│   └── test_socks5.py
├── examples/
│   ├── client.yaml
│   ├── server.yaml
│   ├── client-quic.yaml
│   └── server-quic.yaml
├── docs/
│   ├── QUIC.md
│   ├── HIGH-LOAD-QUIC.md
│   └── PYTHON_GUIDE.md
├── setup.py
├── pyproject.toml
├── requirements.txt
├── requirements-dev.txt
├── README.md
├── LICENSE
└── .gitignore
```

## Implementation Phases

### Phase 1: Foundation (Week 1-2)
- [ ] Set up Python project structure
- [ ] Implement configuration parsing and validation
- [ ] Create logging system
- [ ] Set up CLI framework with Click
- [ ] Write unit tests for configuration

### Phase 2: Network Layer (Week 2-3)
- [ ] Implement raw socket handling with pypcap/scapy
- [ ] Create TCP packet crafting and parsing
- [ ] Implement packet injection and capture
- [ ] Add buffer management
- [ ] Test packet handling on different platforms

### Phase 3: KCP Transport (Week 3-5)
- [ ] Research and select KCP implementation approach
  - Option A: Port KCP algorithm from Go
  - Option B: Use C library with Python bindings
  - Option C: Find/adapt existing Python KCP library
- [ ] Implement KCP connection management
- [ ] Add encryption support (AES, SM4)
- [ ] Implement stream multiplexing
- [ ] Test KCP reliability and performance

### Phase 4: QUIC Transport (Week 5-6)
- [ ] Integrate aioquic library
- [ ] Implement QUIC connection handling
- [ ] Add certificate management
- [ ] Support 0-RTT connections
- [ ] Performance testing and optimization

### Phase 5: Proxy Layer (Week 6-7)
- [ ] Implement SOCKS5 server
- [ ] Add port forwarding support
- [ ] Create connection pooling
- [ ] Handle concurrent connections
- [ ] Test with various SOCKS5 clients

### Phase 6: Client Mode (Week 7-8)
- [ ] Integrate all components for client mode
- [ ] Implement local proxy listening
- [ ] Add transport connection to server
- [ ] Handle connection lifecycle
- [ ] End-to-end testing

### Phase 7: Server Mode (Week 8-9)
- [ ] Implement server listening
- [ ] Handle incoming transport connections
- [ ] Create target forwarding logic
- [ ] Add connection tracking
- [ ] Load testing

### Phase 8: Utilities (Week 9-10)
- [ ] Implement ping command
- [ ] Create dump command with packet analysis
- [ ] Add secret key generation
- [ ] Implement interface detection
- [ ] Version information

### Phase 9: Testing & Documentation (Week 10-12)
- [ ] Comprehensive unit tests
- [ ] Integration tests
- [ ] Performance benchmarks
- [ ] Complete API documentation
- [ ] User guides and examples
- [ ] Migration guide from Go version

### Phase 10: Packaging & Release (Week 12)
- [ ] Create setup.py and pyproject.toml
- [ ] Build distribution packages
- [ ] Set up CI/CD pipelines
- [ ] Prepare release notes
- [ ] Publish to PyPI

## Technical Challenges & Solutions

### 1. Raw Socket Access

**Challenge**: Python needs elevated privileges for raw sockets.

**Solution**: 
- Use `setcap` on Linux: `sudo setcap cap_net_raw,cap_net_admin=eip python3`
- Document privilege requirements clearly
- Consider using libpcap's privilege dropping features

### 2. Performance

**Challenge**: Python is slower than Go, especially for packet processing.

**Solution**:
- Use Cython for performance-critical sections
- Leverage PyPy for JIT compilation
- Consider using C extensions for KCP
- Use asyncio for concurrent operations
- Profile and optimize hot paths

### 3. KCP Implementation

**Challenge**: No mature KCP library for Python.

**Solution**:
- Create Python bindings for C KCP library
- Or port KCP algorithm carefully (reference: https://github.com/skywind3000/kcp)
- Ensure compatibility with Go kcp-go protocol

### 4. Cross-Platform Support

**Challenge**: Different packet capture APIs on Windows/Linux/macOS.

**Solution**:
- Use pypcap or pcapy-ng (supports all platforms)
- Provide platform-specific installation instructions
- Test on all three major platforms

### 5. Async I/O

**Challenge**: Mixing async and sync code, especially with pcap.

**Solution**:
- Use asyncio throughout
- Run blocking pcap operations in executor threads
- Use async/await for all I/O operations

## Performance Considerations

### Optimization Strategies

1. **Use Cython for Critical Paths**
   ```python
   # Compile packet processing to C
   # kcp_processing.pyx
   cdef class KCPProcessor:
       cpdef process_packet(self, bytes data):
           # Fast C-level processing
   ```

2. **Implement Connection Pooling**
   ```python
   from multiprocessing import Pool
   # Use process pool for handling multiple connections
   ```

3. **Buffer Management**
   ```python
   # Pre-allocate buffers to reduce memory allocation overhead
   import array
   buffer_pool = [array.array('B', [0] * 65535) for _ in range(100)]
   ```

4. **Consider PyPy**
   - PyPy's JIT can significantly improve performance
   - Test compatibility with all dependencies

## Testing Strategy

### Unit Tests
```python
# tests/test_kcp.py
import pytest
from paqet.transport.kcp import KCPConnection

def test_kcp_send_recv():
    # Test basic send/receive
    pass

def test_kcp_reliability():
    # Test packet loss recovery
    pass
```

### Integration Tests
```python
# tests/test_integration.py
import pytest
from paqet.client import PaqetClient
from paqet.server import PaqetServer

@pytest.mark.asyncio
async def test_client_server_connection():
    # Test full client-server flow
    pass
```

### Performance Tests
```python
# tests/test_performance.py
import time
from paqet.transport.kcp import KCPConnection

def test_throughput():
    # Measure data transfer rate
    pass

def test_latency():
    # Measure round-trip time
    pass
```

## Dependencies Installation

### requirements.txt
```
# Core dependencies
scapy>=2.5.0
pypcap>=1.2.3
aioquic>=0.9.21
click>=8.1.0
PyYAML>=6.0
cryptography>=41.0.0

# Network
aiohttp>=3.8.0
PySocks>=1.7.1

# Development
pytest>=7.4.0
pytest-asyncio>=0.21.0
pytest-cov>=4.1.0
black>=23.0.0
flake8>=6.0.0
mypy>=1.5.0
```

### requirements-dev.txt
```
# Testing
pytest>=7.4.0
pytest-asyncio>=0.21.0
pytest-cov>=4.1.0
pytest-mock>=3.11.0

# Code quality
black>=23.0.0
flake8>=6.0.0
isort>=5.12.0
mypy>=1.5.0
pylint>=2.17.0

# Documentation
sphinx>=7.1.0
sphinx-rtd-theme>=1.3.0

# Build
build>=0.10.0
twine>=4.0.0
```

## Setup Instructions

### 1. Create Python Project

```bash
# Create new repository
mkdir paqet-python
cd paqet-python

# Initialize git
git init

# Create virtual environment
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Create project structure
mkdir -p paqet/{cli,config,client,server,transport,socket,socks5,protocol,crypto,utils}
mkdir -p tests examples docs

# Create __init__.py files
find paqet -type d -exec touch {}/__init__.py \;
```

### 2. Setup Configuration Files

**setup.py**:
```python
from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setup(
    name="paqet",
    version="0.1.0",
    author="Your Name",
    author_email="your.email@example.com",
    description="Packet-level proxy with KCP/QUIC transport",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/yourusername/paqet-python",
    packages=find_packages(),
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: POSIX :: Linux",
        "Operating System :: MacOS :: MacOS X",
        "Operating System :: Microsoft :: Windows",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
    ],
    python_requires=">=3.8",
    install_requires=[
        "scapy>=2.5.0",
        "pypcap>=1.2.3",
        "aioquic>=0.9.21",
        "click>=8.1.0",
        "PyYAML>=6.0",
        "cryptography>=41.0.0",
        "aiohttp>=3.8.0",
        "PySocks>=1.7.1",
    ],
    entry_points={
        "console_scripts": [
            "paqet=paqet.cli.main:cli",
        ],
    },
)
```

**pyproject.toml**:
```toml
[build-system]
requires = ["setuptools>=45", "wheel", "setuptools_scm[toml]>=6.2"]
build-backend = "setuptools.build_meta"

[project]
name = "paqet"
version = "0.1.0"
description = "Packet-level proxy with KCP/QUIC transport"
readme = "README.md"
requires-python = ">=3.8"
license = {text = "MIT"}
authors = [
    {name = "Your Name", email = "your.email@example.com"}
]
classifiers = [
    "Development Status :: 3 - Alpha",
    "Intended Audience :: Developers",
    "License :: OSI Approved :: MIT License",
    "Programming Language :: Python :: 3",
]
dependencies = [
    "scapy>=2.5.0",
    "pypcap>=1.2.3",
    "aioquic>=0.9.21",
    "click>=8.1.0",
    "PyYAML>=6.0",
    "cryptography>=41.0.0",
]

[project.scripts]
paqet = "paqet.cli.main:cli"

[tool.black]
line-length = 100
target-version = ['py38', 'py39', 'py310', 'py311']

[tool.isort]
profile = "black"
line_length = 100

[tool.mypy]
python_version = "3.8"
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true
```

### 3. Install Development Environment

```bash
# Install in development mode
pip install -e .

# Install development dependencies
pip install -r requirements-dev.txt

# Install system dependencies (Ubuntu/Debian)
sudo apt-get install libpcap-dev python3-dev

# Install system dependencies (macOS)
# libpcap comes with Xcode Command Line Tools
xcode-select --install

# Install system dependencies (Windows)
# Download and install Npcap from https://npcap.com/
```

## Migration Checklist

### Code Structure
- [ ] Port configuration system
- [ ] Port logging system
- [ ] Port packet handling
- [ ] Port KCP transport
- [ ] Port QUIC transport
- [ ] Port SOCKS5 proxy
- [ ] Port client mode
- [ ] Port server mode
- [ ] Port CLI commands
- [ ] Port utility functions

### Testing
- [ ] Port existing Go tests to Python
- [ ] Add platform-specific tests
- [ ] Add performance benchmarks
- [ ] Add integration tests

### Documentation
- [ ] Update README for Python
- [ ] Create Python-specific installation guide
- [ ] Document API differences from Go version
- [ ] Create troubleshooting guide

### Deployment
- [ ] Create GitHub Actions workflow
- [ ] Set up PyPI publishing
- [ ] Create Docker images
- [ ] Build binary distributions (PyInstaller)

## Key Differences from Go Version

### 1. Async/Await vs Goroutines
- **Go**: Uses goroutines and channels
- **Python**: Uses `asyncio` with `async`/`await`

### 2. Type System
- **Go**: Static typing, compile-time checks
- **Python**: Dynamic typing, optional type hints with `mypy`

### 3. Performance
- **Go**: Compiled, faster execution
- **Python**: Interpreted, consider using Cython or PyPy for performance-critical code

### 4. Memory Management
- **Go**: Garbage collected, better performance
- **Python**: Garbage collected, reference counting

### 5. Concurrency Model
- **Go**: CSP (Communicating Sequential Processes)
- **Python**: Async I/O with event loop

## Resources

### Python Libraries Documentation
- [Scapy Documentation](https://scapy.readthedocs.io/)
- [aioquic Documentation](https://aioquic.readthedocs.io/)
- [Click Documentation](https://click.palletsprojects.com/)
- [asyncio Documentation](https://docs.python.org/3/library/asyncio.html)

### Protocol Specifications
- [KCP Protocol](https://github.com/skywind3000/kcp/blob/master/README.en.md)
- [QUIC RFC 9000](https://www.rfc-editor.org/rfc/rfc9000.html)
- [SOCKS5 RFC 1928](https://www.rfc-editor.org/rfc/rfc1928.html)

### Reference Implementations
- [Original Go Implementation](https://github.com/amir0241/paqet)
- [GFW Resist TCP Proxy](https://github.com/GFW-knocker/gfw_resist_tcp_proxy)
- [KCP C Implementation](https://github.com/skywind3000/kcp)

## Next Steps

1. **Review and Approve** this migration plan
2. **Set up Python repository** with initial structure
3. **Start Phase 1** implementation (Foundation)
4. **Iterate** through each phase with testing
5. **Document** progress and learnings
6. **Release** beta version for community testing

## Conclusion

This migration plan provides a comprehensive roadmap for porting paqet from Go to Python. The estimated timeline is 12 weeks for a complete implementation, though this may vary based on team size and KCP implementation approach.

The Python version will maintain feature parity with the Go version while leveraging Python's ecosystem and ease of use. Key considerations include performance optimization, platform compatibility, and maintaining protocol compatibility with the original Go implementation.
