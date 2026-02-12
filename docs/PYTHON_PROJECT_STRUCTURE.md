# Python Project Structure Template for paqet

This document provides a detailed Python project structure for the paqet port, including starter code templates for key modules.

## Directory Structure

```
paqet-python/
├── .github/
│   └── workflows/
│       ├── test.yml              # CI/CD for testing
│       └── publish.yml           # PyPI publishing workflow
├── paqet/                         # Main package directory
│   ├── __init__.py
│   ├── __main__.py               # Entry point for `python -m paqet`
│   │
│   ├── cli/                      # Command-line interface
│   │   ├── __init__.py
│   │   ├── main.py              # Main CLI entry point
│   │   ├── run.py               # Run command
│   │   ├── ping.py              # Ping command
│   │   ├── dump.py              # Packet dump command
│   │   ├── secret.py            # Secret key generation
│   │   ├── iface.py             # Interface listing
│   │   └── version.py           # Version information
│   │
│   ├── config/                   # Configuration management
│   │   ├── __init__.py
│   │   ├── models.py            # Config data models (Pydantic)
│   │   ├── loader.py            # YAML config loader
│   │   ├── validator.py         # Config validation
│   │   ├── kcp.py              # KCP-specific config
│   │   ├── quic.py             # QUIC-specific config
│   │   └── network.py          # Network settings
│   │
│   ├── client/                   # Client mode implementation
│   │   ├── __init__.py
│   │   ├── client.py            # Main client class
│   │   ├── proxy.py             # SOCKS5 proxy handling
│   │   ├── forward.py           # Port forwarding
│   │   └── manager.py           # Connection management
│   │
│   ├── server/                   # Server mode implementation
│   │   ├── __init__.py
│   │   ├── server.py            # Main server class
│   │   ├── handler.py           # Connection handler
│   │   └── forwarder.py         # Target forwarding
│   │
│   ├── transport/                # Transport layer protocols
│   │   ├── __init__.py
│   │   ├── base.py              # Abstract transport interface
│   │   ├── kcp/                 # KCP transport
│   │   │   ├── __init__.py
│   │   │   ├── connection.py   # KCP connection
│   │   │   ├── protocol.py     # KCP protocol implementation
│   │   │   ├── segment.py      # KCP segment handling
│   │   │   └── crypto.py       # KCP encryption
│   │   ├── quic/                # QUIC transport
│   │   │   ├── __init__.py
│   │   │   ├── connection.py   # QUIC connection wrapper
│   │   │   ├── config.py       # QUIC configuration
│   │   │   └── handler.py      # QUIC event handling
│   │   └── multiplexer.py       # Stream multiplexing
│   │
│   ├── socket/                   # Raw socket layer
│   │   ├── __init__.py
│   │   ├── raw_socket.py        # Raw socket management
│   │   ├── pcap_handler.py      # Pcap wrapper
│   │   ├── packet_crafter.py    # TCP packet crafting
│   │   ├── packet_parser.py     # Packet parsing
│   │   └── buffer.py            # Buffer management
│   │
│   ├── socks5/                   # SOCKS5 implementation
│   │   ├── __init__.py
│   │   ├── server.py            # SOCKS5 server
│   │   ├── protocol.py          # SOCKS5 protocol
│   │   ├── auth.py              # Authentication handling
│   │   └── relay.py             # Connection relay
│   │
│   ├── protocol/                 # Protocol definitions
│   │   ├── __init__.py
│   │   └── packet.py            # Packet format definitions
│   │
│   ├── crypto/                   # Cryptography
│   │   ├── __init__.py
│   │   ├── aes.py              # AES encryption
│   │   ├── sm4.py              # SM4 encryption
│   │   └── keyderivation.py    # Key derivation
│   │
│   ├── utils/                    # Utility functions
│   │   ├── __init__.py
│   │   ├── checksum.py         # Checksum calculation
│   │   ├── hash.py             # Hash functions
│   │   ├── pool.py             # Object pooling
│   │   ├── errors.py           # Custom exceptions
│   │   └── platform.py         # Platform-specific utilities
│   │
│   └── logging/                  # Logging configuration
│       ├── __init__.py
│       └── logger.py            # Logger setup
│
├── tests/                        # Test suite
│   ├── __init__.py
│   ├── conftest.py              # Pytest configuration
│   ├── unit/                    # Unit tests
│   │   ├── test_config.py
│   │   ├── test_socket.py
│   │   ├── test_kcp.py
│   │   ├── test_quic.py
│   │   ├── test_socks5.py
│   │   └── test_crypto.py
│   ├── integration/             # Integration tests
│   │   ├── test_client_server.py
│   │   ├── test_proxy.py
│   │   └── test_forwarding.py
│   └── performance/             # Performance tests
│       ├── test_throughput.py
│       └── test_latency.py
│
├── examples/                     # Example configurations
│   ├── client.yaml
│   ├── server.yaml
│   ├── client-quic.yaml
│   ├── server-quic.yaml
│   └── high-load.yaml
│
├── docs/                         # Documentation
│   ├── index.md
│   ├── installation.md
│   ├── configuration.md
│   ├── api/
│   │   └── modules.rst
│   ├── guides/
│   │   ├── quickstart.md
│   │   ├── advanced.md
│   │   └── troubleshooting.md
│   └── migration-from-go.md
│
├── scripts/                      # Utility scripts
│   ├── build.sh                 # Build script
│   ├── test.sh                  # Test runner
│   └── benchmark.py             # Benchmarking
│
├── .gitignore
├── .editorconfig
├── setup.py                      # Setup configuration
├── pyproject.toml               # Modern Python project config
├── requirements.txt             # Production dependencies
├── requirements-dev.txt         # Development dependencies
├── README.md
├── LICENSE
├── CHANGELOG.md
└── MANIFEST.in                  # Package manifest
```

## Starter Code Templates

### 1. Main Entry Point (`paqet/__main__.py`)

```python
"""
Main entry point for running paqet as a module.
Usage: python -m paqet [command] [options]
"""

from paqet.cli.main import cli

if __name__ == "__main__":
    cli()
```

### 2. CLI Main (`paqet/cli/main.py`)

```python
"""Main CLI entry point using Click framework."""

import sys
import click
from paqet import __version__
from paqet.cli import run, ping, dump, secret, iface, version as version_cmd

@click.group()
@click.version_option(version=__version__)
def cli():
    """
    paqet - Packet-level proxy with KCP/QUIC transport
    
    A bidirectional proxy that uses raw sockets to forward traffic
    through encrypted KCP or QUIC transport.
    """
    pass

# Register subcommands
cli.add_command(run.run)
cli.add_command(ping.ping)
cli.add_command(dump.dump)
cli.add_command(secret.secret)
cli.add_command(iface.iface)
cli.add_command(version_cmd.version)

def main():
    """Entry point for console script."""
    try:
        cli()
    except KeyboardInterrupt:
        click.echo("\n\nInterrupted by user", err=True)
        sys.exit(130)
    except Exception as e:
        click.echo(f"Error: {e}", err=True)
        sys.exit(1)

if __name__ == "__main__":
    main()
```

### 3. Run Command (`paqet/cli/run.py`)

```python
"""Run command implementation."""

import asyncio
import click
from pathlib import Path

from paqet.config.loader import load_config
from paqet.client import PaqetClient
from paqet.server import PaqetServer
from paqet.logging import setup_logger

@click.command()
@click.option(
    "-c", "--config",
    type=click.Path(exists=True, path_type=Path),
    required=True,
    help="Path to configuration file"
)
def run(config: Path):
    """Start paqet client or server."""
    
    # Load configuration
    cfg = load_config(config)
    
    # Setup logger
    logger = setup_logger(cfg.log.level)
    
    try:
        if cfg.role == "client":
            logger.info("Starting paqet in client mode")
            client = PaqetClient(cfg)
            asyncio.run(client.start())
        elif cfg.role == "server":
            logger.info("Starting paqet in server mode")
            server = PaqetServer(cfg)
            asyncio.run(server.start())
        else:
            raise ValueError(f"Invalid role: {cfg.role}. Must be 'client' or 'server'")
    except KeyboardInterrupt:
        logger.info("Shutting down...")
    except Exception as e:
        logger.error(f"Failed to start: {e}", exc_info=True)
        raise
```

### 4. Config Models (`paqet/config/models.py`)

```python
"""Configuration data models using Pydantic."""

from typing import Optional, List, Literal
from pydantic import BaseModel, Field, validator
from pathlib import Path

class LogConfig(BaseModel):
    """Logging configuration."""
    level: Literal["none", "debug", "info", "warn", "error", "fatal"] = "info"

class IPv4Config(BaseModel):
    """IPv4 network configuration."""
    addr: str = Field(..., description="IPv4 address and port (e.g., '192.168.1.100:0')")
    router_mac: str = Field(..., description="Gateway MAC address")
    
    @validator("router_mac")
    def validate_mac(cls, v):
        """Validate MAC address format."""
        parts = v.split(":")
        if len(parts) != 6:
            raise ValueError("MAC address must have 6 parts")
        for part in parts:
            if len(part) != 2:
                raise ValueError("Each MAC address part must be 2 hex digits")
        return v.lower()

class NetworkConfig(BaseModel):
    """Network interface configuration."""
    interface: str = Field(..., description="Network interface name")
    guid: Optional[str] = Field(None, description="Windows Npcap device GUID")
    ipv4: IPv4Config
    
    @validator("interface")
    def validate_interface(cls, v):
        """Validate interface name is not empty."""
        if not v.strip():
            raise ValueError("Interface name cannot be empty")
        return v

class TCPFlagsConfig(BaseModel):
    """TCP flags configuration."""
    local_flag: List[str] = Field(default_factory=lambda: ["PA"])
    remote_flag: List[str] = Field(default_factory=lambda: ["PA"])

class KCPConfig(BaseModel):
    """KCP transport configuration."""
    block: str = Field(default="aes", description="Encryption algorithm")
    key: str = Field(..., description="Encryption key")
    data_shard: int = Field(default=10, ge=0)
    parity_shard: int = Field(default=3, ge=0)
    
class QUICConfig(BaseModel):
    """QUIC transport configuration."""
    cert: str = Field(..., description="Certificate file path")
    key: str = Field(..., description="Private key file path")
    ca: Optional[str] = Field(None, description="CA certificate file path")
    
class TransportConfig(BaseModel):
    """Transport layer configuration."""
    protocol: Literal["kcp", "quic"] = "kcp"
    kcp: Optional[KCPConfig] = None
    quic: Optional[QUICConfig] = None
    
    @validator("kcp")
    def validate_kcp(cls, v, values):
        """Ensure KCP config exists when protocol is kcp."""
        if values.get("protocol") == "kcp" and v is None:
            raise ValueError("KCP configuration required when protocol is 'kcp'")
        return v
    
    @validator("quic")
    def validate_quic(cls, v, values):
        """Ensure QUIC config exists when protocol is quic."""
        if values.get("protocol") == "quic" and v is None:
            raise ValueError("QUIC configuration required when protocol is 'quic'")
        return v

class SOCKS5Config(BaseModel):
    """SOCKS5 proxy configuration."""
    listen: str = Field(..., description="Listen address (e.g., '127.0.0.1:1080')")

class ForwardConfig(BaseModel):
    """Port forwarding configuration."""
    listen: str = Field(..., description="Local listen address")
    target: str = Field(..., description="Target address")
    protocol: Literal["tcp", "udp"] = "tcp"

class ServerConnectionConfig(BaseModel):
    """Server connection configuration (client mode)."""
    addr: str = Field(..., description="Server address and port")

class ListenConfig(BaseModel):
    """Listen configuration (server mode)."""
    addr: str = Field(..., description="Listen address and port")

class Config(BaseModel):
    """Main paqet configuration."""
    role: Literal["client", "server"] = Field(..., description="Operating mode")
    log: LogConfig = Field(default_factory=LogConfig)
    network: NetworkConfig
    transport: TransportConfig
    
    # Client-specific
    socks5: Optional[List[SOCKS5Config]] = None
    forward: Optional[List[ForwardConfig]] = None
    server: Optional[ServerConnectionConfig] = None
    
    # Server-specific
    listen: Optional[ListenConfig] = None
    
    @validator("server")
    def validate_server_config(cls, v, values):
        """Ensure server config exists in client mode."""
        if values.get("role") == "client" and v is None:
            raise ValueError("Server configuration required in client mode")
        return v
    
    @validator("listen")
    def validate_listen_config(cls, v, values):
        """Ensure listen config exists in server mode."""
        if values.get("role") == "server" and v is None:
            raise ValueError("Listen configuration required in server mode")
        return v

    class Config:
        """Pydantic config."""
        extra = "forbid"  # Disallow extra fields
        validate_assignment = True
```

### 5. Config Loader (`paqet/config/loader.py`)

```python
"""Configuration file loader."""

import yaml
from pathlib import Path
from typing import Union
from paqet.config.models import Config

def load_config(path: Union[str, Path]) -> Config:
    """
    Load and validate configuration from YAML file.
    
    Args:
        path: Path to configuration file
        
    Returns:
        Validated Config object
        
    Raises:
        FileNotFoundError: If config file doesn't exist
        ValueError: If config is invalid
    """
    path = Path(path)
    
    if not path.exists():
        raise FileNotFoundError(f"Configuration file not found: {path}")
    
    with open(path, "r") as f:
        data = yaml.safe_load(f)
    
    try:
        config = Config(**data)
    except Exception as e:
        raise ValueError(f"Invalid configuration: {e}")
    
    return config
```

### 6. Raw Socket Handler (`paqet/socket/raw_socket.py`)

```python
"""Raw socket handling with pcap."""

import pcap
import struct
import logging
from typing import Optional, Callable
from scapy.all import Ether, IP, TCP, Raw

logger = logging.getLogger(__name__)

class RawSocket:
    """Raw socket handler using libpcap."""
    
    def __init__(self, interface: str, filter_expr: Optional[str] = None):
        """
        Initialize raw socket.
        
        Args:
            interface: Network interface name
            filter_expr: BPF filter expression
        """
        self.interface = interface
        self.pcap = pcap.pcap(
            name=interface,
            snaplen=65535,
            promisc=True,
            immediate=True
        )
        
        if filter_expr:
            self.pcap.setfilter(filter_expr)
        
        logger.info(f"Opened raw socket on {interface}")
    
    def send_packet(self, packet: bytes) -> int:
        """
        Inject raw packet to interface.
        
        Args:
            packet: Raw packet bytes
            
        Returns:
            Number of bytes sent
        """
        try:
            sent = self.pcap.inject(packet)
            logger.debug(f"Sent {sent} bytes")
            return sent
        except Exception as e:
            logger.error(f"Failed to send packet: {e}")
            raise
    
    def recv_packet(self, timeout: int = 1000) -> Optional[tuple[int, bytes]]:
        """
        Receive packet from interface.
        
        Args:
            timeout: Timeout in milliseconds
            
        Returns:
            Tuple of (timestamp, packet_bytes) or None
        """
        try:
            for ts, pkt in self.pcap.readpkts():
                return (ts, pkt)
        except Exception as e:
            logger.error(f"Failed to receive packet: {e}")
        return None
    
    def start_capture(self, callback: Callable[[int, bytes], None]):
        """
        Start continuous packet capture.
        
        Args:
            callback: Function to call for each packet (timestamp, packet_bytes)
        """
        logger.info(f"Starting packet capture on {self.interface}")
        try:
            self.pcap.loop(-1, callback)
        except KeyboardInterrupt:
            logger.info("Packet capture interrupted")
        except Exception as e:
            logger.error(f"Capture error: {e}")
            raise
    
    def close(self):
        """Close the raw socket."""
        logger.info(f"Closing raw socket on {self.interface}")
        # pcap objects don't need explicit closing in most implementations
        pass
```

### 7. TCP Packet Crafter (`paqet/socket/packet_crafter.py`)

```python
"""TCP packet crafting using Scapy."""

from scapy.all import Ether, IP, TCP, Raw
from scapy.layers.inet import checksum
import logging

logger = logging.getLogger(__name__)

class TCPPacketCrafter:
    """TCP packet crafting and manipulation."""
    
    @staticmethod
    def craft_tcp_packet(
        src_mac: str,
        dst_mac: str,
        src_ip: str,
        dst_ip: str,
        src_port: int,
        dst_port: int,
        payload: bytes = b"",
        flags: str = "PA",
        seq: int = 0,
        ack: int = 0,
        window: int = 65535
    ) -> bytes:
        """
        Craft a TCP packet with specified parameters.
        
        Args:
            src_mac: Source MAC address
            dst_mac: Destination MAC address
            src_ip: Source IP address
            dst_ip: Destination IP address
            src_port: Source port
            dst_port: Destination port
            payload: Packet payload
            flags: TCP flags (e.g., "S", "PA", "A", "F")
            seq: Sequence number
            ack: Acknowledgment number
            window: TCP window size
            
        Returns:
            Raw packet bytes
        """
        # Build packet layers
        eth = Ether(src=src_mac, dst=dst_mac)
        ip = IP(src=src_ip, dst=dst_ip)
        tcp = TCP(
            sport=src_port,
            dport=dst_port,
            flags=flags,
            seq=seq,
            ack=ack,
            window=window
        )
        
        # Add payload if present
        if payload:
            packet = eth / ip / tcp / Raw(load=payload)
        else:
            packet = eth / ip / tcp
        
        # Let Scapy calculate checksums
        del packet[IP].chksum
        del packet[TCP].chksum
        
        # Return raw bytes
        return bytes(packet)
    
    @staticmethod
    def parse_tcp_packet(packet_bytes: bytes) -> dict:
        """
        Parse TCP packet and extract information.
        
        Args:
            packet_bytes: Raw packet bytes
            
        Returns:
            Dictionary with packet information
        """
        from scapy.all import Ether
        
        try:
            packet = Ether(packet_bytes)
            
            if not packet.haslayer(TCP):
                return None
            
            info = {
                "src_mac": packet[Ether].src,
                "dst_mac": packet[Ether].dst,
                "src_ip": packet[IP].src,
                "dst_ip": packet[IP].dst,
                "src_port": packet[TCP].sport,
                "dst_port": packet[TCP].dport,
                "flags": packet[TCP].flags,
                "seq": packet[TCP].seq,
                "ack": packet[TCP].ack,
                "window": packet[TCP].window,
                "payload": bytes(packet[TCP].payload) if packet[TCP].payload else b""
            }
            
            return info
        except Exception as e:
            logger.error(f"Failed to parse packet: {e}")
            return None
```

### 8. Client Implementation (`paqet/client/client.py`)

```python
"""Paqet client implementation."""

import asyncio
import logging
from typing import Optional

from paqet.config.models import Config
from paqet.socket.raw_socket import RawSocket
from paqet.transport.kcp import KCPTransport
from paqet.transport.quic import QUICTransport
from paqet.socks5.server import SOCKS5Server

logger = logging.getLogger(__name__)

class PaqetClient:
    """Paqet client - SOCKS5 proxy mode."""
    
    def __init__(self, config: Config):
        """
        Initialize paqet client.
        
        Args:
            config: Client configuration
        """
        if config.role != "client":
            raise ValueError("Config role must be 'client'")
        
        self.config = config
        self.raw_socket: Optional[RawSocket] = None
        self.transport = None
        self.socks5_servers = []
        self.running = False
        
        logger.info("Paqet client initialized")
    
    def _init_transport(self):
        """Initialize transport layer based on configuration."""
        if self.config.transport.protocol == "kcp":
            logger.info("Using KCP transport")
            self.transport = KCPTransport(
                self.config.transport.kcp,
                self.raw_socket
            )
        elif self.config.transport.protocol == "quic":
            logger.info("Using QUIC transport")
            self.transport = QUICTransport(
                self.config.transport.quic,
                self.raw_socket
            )
    
    async def start(self):
        """Start the client proxy."""
        try:
            logger.info("Starting paqet client")
            
            # Initialize raw socket
            self.raw_socket = RawSocket(
                self.config.network.interface,
                filter_expr=None  # Will be set based on config
            )
            
            # Initialize transport
            self._init_transport()
            
            # Connect to server
            await self._connect_to_server()
            
            # Start SOCKS5 servers
            await self._start_socks5_servers()
            
            # Start packet processing
            self.running = True
            await self._process_packets()
            
        except Exception as e:
            logger.error(f"Client error: {e}", exc_info=True)
            raise
        finally:
            await self.stop()
    
    async def _connect_to_server(self):
        """Connect to paqet server."""
        logger.info(f"Connecting to server {self.config.server.addr}")
        # TODO: Implement server connection
        await asyncio.sleep(0)  # Placeholder
    
    async def _start_socks5_servers(self):
        """Start SOCKS5 proxy servers."""
        if not self.config.socks5:
            logger.warning("No SOCKS5 configuration found")
            return
        
        for socks_config in self.config.socks5:
            host, port = socks_config.listen.rsplit(":", 1)
            server = SOCKS5Server(host, int(port), self.transport)
            self.socks5_servers.append(server)
            asyncio.create_task(server.start())
            logger.info(f"SOCKS5 server listening on {socks_config.listen}")
    
    async def _process_packets(self):
        """Process incoming raw packets."""
        logger.info("Starting packet processing")
        while self.running:
            # TODO: Implement packet processing loop
            await asyncio.sleep(0.1)
    
    async def stop(self):
        """Stop the client."""
        logger.info("Stopping paqet client")
        self.running = False
        
        # Stop SOCKS5 servers
        for server in self.socks5_servers:
            await server.stop()
        
        # Close raw socket
        if self.raw_socket:
            self.raw_socket.close()
        
        logger.info("Paqet client stopped")
```

### 9. Setup.py

```python
"""Setup configuration for paqet."""

from setuptools import setup, find_packages
from pathlib import Path

# Read README
readme_file = Path(__file__).parent / "README.md"
long_description = readme_file.read_text(encoding="utf-8") if readme_file.exists() else ""

# Read version
version_file = Path(__file__).parent / "paqet" / "__init__.py"
version = "0.1.0"
if version_file.exists():
    for line in version_file.read_text().splitlines():
        if line.startswith("__version__"):
            version = line.split("=")[1].strip().strip('"').strip("'")
            break

setup(
    name="paqet",
    version=version,
    author="Amir",
    description="Packet-level proxy with KCP/QUIC transport",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/yourusername/paqet-python",
    packages=find_packages(exclude=["tests", "tests.*"]),
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "Intended Audience :: System Administrators",
        "License :: OSI Approved :: MIT License",
        "Operating System :: POSIX :: Linux",
        "Operating System :: MacOS :: MacOS X",
        "Operating System :: Microsoft :: Windows",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Internet :: Proxy Servers",
        "Topic :: System :: Networking",
    ],
    python_requires=">=3.8",
    install_requires=[
        "scapy>=2.5.0",
        "pypcap>=1.2.3",
        "aioquic>=0.9.21",
        "click>=8.1.0",
        "PyYAML>=6.0",
        "pydantic>=2.0.0",
        "cryptography>=41.0.0",
        "aiohttp>=3.8.0",
        "PySocks>=1.7.1",
        "netifaces>=0.11.0",
        "structlog>=23.1.0",
    ],
    extras_require={
        "dev": [
            "pytest>=7.4.0",
            "pytest-asyncio>=0.21.0",
            "pytest-cov>=4.1.0",
            "pytest-mock>=3.11.0",
            "black>=23.0.0",
            "flake8>=6.0.0",
            "isort>=5.12.0",
            "mypy>=1.5.0",
            "pylint>=2.17.0",
        ],
        "docs": [
            "sphinx>=7.1.0",
            "sphinx-rtd-theme>=1.3.0",
            "sphinx-click>=5.0.0",
        ],
    },
    entry_points={
        "console_scripts": [
            "paqet=paqet.cli.main:main",
        ],
    },
    include_package_data=True,
    zip_safe=False,
)
```

### 10. PyProject.toml

```toml
[build-system]
requires = ["setuptools>=45", "wheel", "setuptools_scm>=6.2"]
build-backend = "setuptools.build_meta"

[project]
name = "paqet"
version = "0.1.0"
description = "Packet-level proxy with KCP/QUIC transport"
readme = "README.md"
requires-python = ">=3.8"
license = {text = "MIT"}
authors = [
    {name = "Amir", email = "your.email@example.com"}
]
keywords = ["proxy", "kcp", "quic", "packet", "raw-socket", "vpn"]
classifiers = [
    "Development Status :: 3 - Alpha",
    "Intended Audience :: Developers",
    "License :: OSI Approved :: MIT License",
    "Programming Language :: Python :: 3",
    "Programming Language :: Python :: 3.8",
    "Programming Language :: Python :: 3.9",
    "Programming Language :: Python :: 3.10",
    "Programming Language :: Python :: 3.11",
]

dependencies = [
    "scapy>=2.5.0",
    "pypcap>=1.2.3",
    "aioquic>=0.9.21",
    "click>=8.1.0",
    "PyYAML>=6.0",
    "pydantic>=2.0.0",
    "cryptography>=41.0.0",
    "aiohttp>=3.8.0",
    "netifaces>=0.11.0",
]

[project.optional-dependencies]
dev = [
    "pytest>=7.4.0",
    "pytest-asyncio>=0.21.0",
    "pytest-cov>=4.1.0",
    "black>=23.0.0",
    "mypy>=1.5.0",
]

[project.scripts]
paqet = "paqet.cli.main:main"

[project.urls]
Homepage = "https://github.com/yourusername/paqet-python"
Documentation = "https://paqet-python.readthedocs.io"
Repository = "https://github.com/yourusername/paqet-python"
Issues = "https://github.com/yourusername/paqet-python/issues"

[tool.black]
line-length = 100
target-version = ['py38', 'py39', 'py310', 'py311']
include = '\.pyi?$'

[tool.isort]
profile = "black"
line_length = 100

[tool.mypy]
python_version = "3.8"
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = false
ignore_missing_imports = true

[tool.pytest.ini_options]
testpaths = ["tests"]
python_files = ["test_*.py"]
python_classes = ["Test*"]
python_functions = ["test_*"]
addopts = "-v --cov=paqet --cov-report=term-missing"
asyncio_mode = "auto"

[tool.coverage.run]
source = ["paqet"]
omit = ["tests/*", "*/tests/*", "setup.py"]

[tool.coverage.report]
exclude_lines = [
    "pragma: no cover",
    "def __repr__",
    "raise AssertionError",
    "raise NotImplementedError",
    "if __name__ == .__main__.:",
    "if TYPE_CHECKING:",
]
```

## Quick Start Script

Create a script to set up the Python project structure:

```bash
#!/bin/bash
# setup_python_project.sh - Initialize paqet Python project

set -e

PROJECT_NAME="paqet-python"
echo "Creating Python project structure for $PROJECT_NAME"

# Create main directory
mkdir -p $PROJECT_NAME
cd $PROJECT_NAME

# Create package directories
mkdir -p paqet/{cli,config,client,server,transport/{kcp,quic},socket,socks5,protocol,crypto,utils,logging}
mkdir -p tests/{unit,integration,performance}
mkdir -p examples docs/guides scripts

# Create __init__.py files
find paqet -type d -exec touch {}/__init__.py \;
touch tests/__init__.py

# Create main __init__.py with version
cat > paqet/__init__.py << 'EOF'
"""
paqet - Packet-level proxy with KCP/QUIC transport.
"""

__version__ = "0.1.0"
__author__ = "Amir"
__license__ = "MIT"

from paqet.config.loader import load_config
from paqet.client import PaqetClient
from paqet.server import PaqetServer

__all__ = ["load_config", "PaqetClient", "PaqetServer"]
EOF

# Create basic files
touch README.md LICENSE CHANGELOG.md
touch requirements.txt requirements-dev.txt
touch setup.py pyproject.toml
touch .gitignore .editorconfig

# Create .gitignore
cat > .gitignore << 'EOF'
# Python
__pycache__/
*.py[cod]
*$py.class
*.so
.Python
build/
develop-eggs/
dist/
downloads/
eggs/
.eggs/
lib/
lib64/
parts/
sdist/
var/
wheels/
*.egg-info/
.installed.cfg
*.egg
MANIFEST

# Virtual environments
venv/
ENV/
env/
.venv

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# Testing
.pytest_cache/
.coverage
htmlcov/
.tox/

# Logs
*.log

# OS
.DS_Store
Thumbs.db
EOF

echo "✓ Python project structure created successfully!"
echo ""
echo "Next steps:"
echo "1. cd $PROJECT_NAME"
echo "2. python3 -m venv venv"
echo "3. source venv/bin/activate"
echo "4. pip install -e ."
echo "5. Start implementing modules from the migration guide"
```

## Next Steps

1. **Run the setup script** to create the project structure
2. **Implement modules incrementally** following the migration guide
3. **Write tests** for each module as you implement it
4. **Document** APIs and usage as you go
5. **Test on multiple platforms** (Linux, macOS, Windows)

## Resources

- [Go Implementation](https://github.com/amir0241/paqet)
- [Python Migration Guide](./PYTHON_MIGRATION.md)
- [Scapy Documentation](https://scapy.readthedocs.io/)
- [aioquic Documentation](https://aioquic.readthedocs.io/)
- [Click Documentation](https://click.palletsprojects.com/)
