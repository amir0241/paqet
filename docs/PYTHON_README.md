# Python Version Roadmap for paqet

This document provides a comprehensive roadmap for creating a full Python version of the paqet project.

## Executive Summary

**Project**: paqet - Transport over raw packets  
**Current Language**: Go 1.25+  
**Target Language**: Python 3.8+  
**Estimated Timeline**: 10-12 weeks  
**Lines of Code**: ~5,685 (Go) → Estimated ~7,000-8,000 (Python)  

## What is paqet?

paqet is a bidirectional packet-level proxy that uses raw sockets to forward traffic. It captures packets using pcap and injects crafted TCP packets containing encrypted transport data using either KCP or QUIC protocols.

### Key Features

- ✅ Raw packet capture and injection using libpcap
- ✅ KCP transport protocol (reliable UDP with encryption)
- ✅ QUIC transport protocol (modern IETF standard)
- ✅ SOCKS5 proxy server
- ✅ Port forwarding
- ✅ Stream multiplexing
- ✅ Multiple encryption algorithms (AES, SM4)
- ✅ Cross-platform support (Linux, macOS, Windows)

## Why Python?

### Advantages
1. **Rapid Development**: Faster prototyping and iteration
2. **Rich Ecosystem**: Extensive networking and cryptography libraries
3. **Easier Maintenance**: More readable code, easier to contribute
4. **Better Debugging**: Dynamic nature makes debugging easier
5. **Integration**: Easier to integrate with Python-based tools and workflows

### Challenges
1. **Performance**: Python is slower than Go for packet processing
2. **Concurrency**: Different model (asyncio vs goroutines)
3. **Packaging**: More complex distribution (though improving)
4. **Dependencies**: More external dependencies to manage

## Technology Mapping

### Core Dependencies

| Component | Go | Python | Status |
|-----------|-----|---------|--------|
| Packet Handling | gopacket | scapy | ✅ Mature |
| Raw Sockets | libpcap | pypcap | ✅ Available |
| KCP Protocol | kcp-go | Custom/Bindings | ⚠️ Need implementation |
| QUIC Protocol | quic-go | aioquic | ✅ Mature |
| Multiplexing | smux | Custom | ⚠️ Need implementation |
| CLI Framework | cobra | click | ✅ Mature |
| Config | go-yaml | PyYAML + Pydantic | ✅ Mature |
| Crypto | Go crypto | cryptography | ✅ Mature |
| SOCKS5 | txthinking/socks5 | Custom/PySocks | ✅ Available |

## Project Structure

```
paqet-python/
├── paqet/                        # Main package
│   ├── cli/                     # Command-line interface
│   ├── config/                  # Configuration management
│   ├── client/                  # Client mode
│   ├── server/                  # Server mode
│   ├── transport/               # Transport protocols
│   │   ├── kcp/                # KCP implementation
│   │   └── quic/               # QUIC wrapper
│   ├── socket/                  # Raw socket handling
│   ├── socks5/                  # SOCKS5 server
│   ├── protocol/                # Protocol definitions
│   ├── crypto/                  # Encryption
│   └── utils/                   # Utilities
├── tests/                       # Test suite
├── examples/                    # Example configs
├── docs/                        # Documentation
└── scripts/                     # Utility scripts
```

## Implementation Phases

### Phase 1: Foundation (Weeks 1-2)
**Goal**: Set up project infrastructure and basic components

- [x] Create comprehensive migration guide
- [x] Document Python dependencies and mapping
- [x] Design Python project structure
- [ ] Set up Python project with proper packaging
- [ ] Implement configuration system with Pydantic
- [ ] Create logging infrastructure
- [ ] Set up CLI framework with Click
- [ ] Write unit tests for configuration

**Deliverables**:
- Working Python project setup
- Config parsing and validation
- Basic CLI structure
- Test infrastructure

### Phase 2: Network Layer (Weeks 2-3)
**Goal**: Implement raw socket and packet handling

- [ ] Implement raw socket wrapper using pypcap
- [ ] Create TCP packet crafting with Scapy
- [ ] Implement packet parsing and validation
- [ ] Add checksum calculation
- [ ] Create buffer management system
- [ ] Platform-specific handling (Linux/macOS/Windows)
- [ ] Write tests for packet operations

**Deliverables**:
- Working raw socket I/O
- TCP packet creation and parsing
- Cross-platform compatibility

### Phase 3: KCP Transport (Weeks 3-5)
**Goal**: Implement or integrate KCP protocol

**Decision Point**: Choose KCP implementation approach
1. **Option A**: Port KCP from C to Python (most compatible)
2. **Option B**: Create Python bindings to C KCP library (fastest)
3. **Option C**: Find/adapt existing Python KCP (if available)

Tasks:
- [ ] Research and select KCP approach
- [ ] Implement KCP connection management
- [ ] Add KCP segment handling
- [ ] Implement congestion control
- [ ] Add encryption layer (AES, SM4)
- [ ] Create stream multiplexing
- [ ] Performance testing and optimization
- [ ] Comprehensive KCP tests

**Deliverables**:
- Working KCP transport
- Encryption support
- Performance benchmarks

### Phase 4: QUIC Transport (Weeks 5-6)
**Goal**: Integrate QUIC using aioquic

- [ ] Wrapper around aioquic library
- [ ] QUIC connection management
- [ ] Certificate handling
- [ ] 0-RTT support
- [ ] Stream multiplexing for QUIC
- [ ] Performance optimization
- [ ] QUIC-specific tests

**Deliverables**:
- Working QUIC transport
- TLS certificate management
- Performance comparison with KCP

### Phase 5: Proxy Layer (Weeks 6-7)
**Goal**: Implement SOCKS5 and port forwarding

- [ ] SOCKS5 protocol implementation
- [ ] Authentication handling
- [ ] Connection relay logic
- [ ] Port forwarding support
- [ ] Connection pooling
- [ ] Error handling and timeouts
- [ ] Proxy tests

**Deliverables**:
- Working SOCKS5 server
- Port forwarding functionality
- Connection management

### Phase 6: Client Mode (Weeks 7-8)
**Goal**: Complete client implementation

- [ ] Integrate all components for client
- [ ] Local proxy listener
- [ ] Connection to remote server
- [ ] Request routing
- [ ] Connection lifecycle management
- [ ] Error recovery
- [ ] End-to-end client tests

**Deliverables**:
- Fully functional client mode
- Tested with various applications

### Phase 7: Server Mode (Weeks 8-9)
**Goal**: Complete server implementation

- [ ] Server listening logic
- [ ] Handle incoming connections
- [ ] Target forwarding
- [ ] Connection tracking
- [ ] Resource management
- [ ] Load testing
- [ ] Server-side tests

**Deliverables**:
- Fully functional server mode
- Load tested and optimized

### Phase 8: CLI Commands (Weeks 9-10)
**Goal**: Implement utility commands

- [ ] `paqet run` - Main command
- [ ] `paqet ping` - Connectivity test
- [ ] `paqet dump` - Packet analysis
- [ ] `paqet secret` - Key generation
- [ ] `paqet iface` - Interface listing
- [ ] `paqet version` - Version info
- [ ] Command tests

**Deliverables**:
- Complete CLI interface
- All utility commands working

### Phase 9: Testing & Documentation (Weeks 10-11)
**Goal**: Comprehensive testing and docs

- [ ] Unit test coverage (>80%)
- [ ] Integration tests
- [ ] Performance benchmarks
- [ ] Platform compatibility tests
- [ ] API documentation
- [ ] User guide
- [ ] Migration guide from Go
- [ ] Troubleshooting guide

**Deliverables**:
- Complete test suite
- Comprehensive documentation
- Performance analysis

### Phase 10: Packaging & Release (Week 12)
**Goal**: Prepare for release

- [ ] Package for PyPI
- [ ] Create wheels for major platforms
- [ ] Set up CI/CD (GitHub Actions)
- [ ] Create Docker images
- [ ] Binary distributions (PyInstaller)
- [ ] Release notes
- [ ] Public announcement

**Deliverables**:
- PyPI package
- Distribution packages
- Release v0.1.0

## Installation Guide

### Prerequisites

**System Requirements**:
- Python 3.8 or higher
- libpcap development libraries
- C compiler (for some dependencies)
- Root/Administrator privileges (for raw sockets)

**Platform-specific**:

**Linux (Debian/Ubuntu)**:
```bash
sudo apt-get update
sudo apt-get install python3 python3-pip python3-dev
sudo apt-get install libpcap-dev gcc
```

**Linux (RHEL/CentOS/Fedora)**:
```bash
sudo yum install python3 python3-pip python3-devel
sudo yum install libpcap-devel gcc
```

**macOS**:
```bash
# Install Xcode Command Line Tools (includes libpcap)
xcode-select --install

# Install Python (if using Homebrew)
brew install python
```

**Windows**:
1. Install Python from [python.org](https://www.python.org/downloads/)
2. Install [Npcap](https://npcap.com/)
3. Install Visual C++ Build Tools

### Python Environment Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/paqet-python.git
cd paqet-python

# Create virtual environment
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Upgrade pip
pip install --upgrade pip

# Install paqet in development mode
pip install -e .

# Or install from PyPI (once released)
pip install paqet

# Verify installation
paqet version
```

### Granting Raw Socket Permissions

**Linux**:
```bash
# Option 1: Run with sudo (simple but requires password)
sudo paqet run -c config.yaml

# Option 2: Grant capabilities (more secure)
sudo setcap cap_net_raw,cap_net_admin=eip $(which python3)
paqet run -c config.yaml
```

**macOS**:
```bash
# Must run with sudo
sudo paqet run -c config.yaml
```

**Windows**:
```bash
# Run PowerShell or CMD as Administrator
paqet run -c config.yaml
```

## Configuration

Configuration is identical to the Go version, using YAML format. The Python version maintains full compatibility with existing config files.

### Example Client Configuration

```yaml
role: "client"

log:
  level: "info"

socks5:
  - listen: "127.0.0.1:1080"

network:
  interface: "en0"
  ipv4:
    addr: "192.168.1.100:0"
    router_mac: "aa:bb:cc:dd:ee:ff"

server:
  addr: "10.0.0.100:9999"

transport:
  protocol: "kcp"
  kcp:
    block: "aes"
    key: "your-secret-key-here"
```

### Example Server Configuration

```yaml
role: "server"

log:
  level: "info"

listen:
  addr: ":9999"

network:
  interface: "eth0"
  ipv4:
    addr: "10.0.0.100:9999"
    router_mac: "aa:bb:cc:dd:ee:ff"

transport:
  protocol: "kcp"
  kcp:
    block: "aes"
    key: "your-secret-key-here"
```

## Usage Examples

### Start Client

```bash
# With sudo (Linux/macOS)
sudo paqet run -c client.yaml

# As Administrator (Windows)
paqet run -c client.yaml
```

### Start Server

```bash
# With sudo (Linux)
sudo paqet run -c server.yaml
```

### Test Connection

```bash
# Using curl with SOCKS5 proxy
curl -v https://httpbin.org/ip --proxy socks5h://127.0.0.1:1080

# Using Python requests
import requests
proxies = {
    'http': 'socks5://127.0.0.1:1080',
    'https': 'socks5://127.0.0.1:1080'
}
response = requests.get('https://httpbin.org/ip', proxies=proxies)
print(response.json())
```

### Generate Secret Key

```bash
paqet secret
# Output: Generated secret key: a1b2c3d4e5f6...
```

### Packet Dump

```bash
# Monitor packets on port 9999
sudo paqet dump -p 9999 -i eth0
```

### Ping Server

```bash
# Test connectivity
sudo paqet ping -c client.yaml
```

## Performance Considerations

### Optimization Strategies

1. **Use PyPy**: PyPy's JIT compiler can significantly improve performance
   ```bash
   pypy3 -m pip install paqet
   sudo pypy3 -m paqet run -c config.yaml
   ```

2. **Cython Compilation**: Compile performance-critical modules
   ```bash
   pip install cython
   python setup.py build_ext --inplace
   ```

3. **Profile and Optimize**:
   ```bash
   python -m cProfile -o profile.stats -m paqet run -c config.yaml
   python -m pstats profile.stats
   ```

### Expected Performance

| Metric | Go Version | Python (CPython) | Python (PyPy) |
|--------|-----------|------------------|---------------|
| Throughput | 100% | 30-50% | 60-80% |
| Latency | 1x | 1.5-2x | 1.2-1.5x |
| Memory | 1x | 1.5-2x | 1.2-1.8x |
| CPU | 1x | 2-3x | 1.5-2x |

*Note: These are estimates. Actual performance may vary.*

### When to Use Python vs Go

**Use Python version when**:
- You need rapid development and iteration
- You want easier debugging and testing
- You're integrating with Python tools/workflows
- Performance is acceptable for your use case
- You value code readability and maintainability

**Use Go version when**:
- You need maximum performance
- You're handling high packet rates
- You want lower resource usage
- You need the smallest binary size
- You want faster startup time

## Development Guide

### Setting Up Development Environment

```bash
# Clone and setup
git clone https://github.com/yourusername/paqet-python.git
cd paqet-python
python3 -m venv venv
source venv/bin/activate

# Install in development mode with dev dependencies
pip install -e ".[dev]"

# Install pre-commit hooks
pip install pre-commit
pre-commit install
```

### Running Tests

```bash
# Run all tests
pytest

# Run with coverage
pytest --cov=paqet --cov-report=html

# Run specific test file
pytest tests/unit/test_config.py

# Run tests in parallel
pytest -n auto
```

### Code Style

```bash
# Format code
black paqet/ tests/
isort paqet/ tests/

# Lint code
flake8 paqet/ tests/
pylint paqet/

# Type check
mypy paqet/
```

### Building Documentation

```bash
cd docs/
sphinx-build -b html . _build/html
# Open _build/html/index.html in browser
```

## Migration from Go Version

### Key Differences

1. **Async/Await vs Goroutines**:
   ```python
   # Python (asyncio)
   async def handle_connection(reader, writer):
       data = await reader.read(1024)
       writer.write(response)
       await writer.drain()
   
   # Go (goroutines)
   go func() {
       data := make([]byte, 1024)
       conn.Read(data)
       conn.Write(response)
   }()
   ```

2. **Error Handling**:
   ```python
   # Python (exceptions)
   try:
       result = operation()
   except Exception as e:
       logger.error(f"Operation failed: {e}")
   
   # Go (error returns)
   result, err := operation()
   if err != nil {
       log.Errorf("Operation failed: %v", err)
   }
   ```

3. **Type System**:
   ```python
   # Python (optional type hints)
   def process_packet(data: bytes) -> Optional[dict]:
       ...
   
   # Go (static types)
   func processPacket(data []byte) (*PacketInfo, error) {
       ...
   }
   ```

### Configuration Compatibility

The Python version maintains 100% configuration compatibility with the Go version. You can use the same YAML files for both.

### Protocol Compatibility

The Python implementation follows the same protocols as the Go version:
- KCP packet format is identical
- QUIC uses standard IETF QUIC
- TCP packet crafting is the same

This means Python clients can connect to Go servers and vice versa.

## Troubleshooting

### Common Issues

**1. Permission Denied**
```bash
# Error: Operation not permitted
# Solution: Run with sudo or grant capabilities
sudo paqet run -c config.yaml
```

**2. Module Import Errors**
```bash
# Error: No module named 'pcap'
# Solution: Install libpcap-dev and reinstall pypcap
sudo apt-get install libpcap-dev
pip install --no-cache-dir pypcap
```

**3. QUIC Certificate Errors**
```bash
# Error: Certificate verification failed
# Solution: Generate proper certificates or disable verification
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
```

**4. Performance Issues**
```bash
# Try PyPy for better performance
pip install pypy3
pypy3 -m pip install paqet
sudo pypy3 -m paqet run -c config.yaml
```

### Debug Mode

```bash
# Enable debug logging
# In config.yaml:
log:
  level: "debug"

# Or via environment
PAQET_LOG_LEVEL=debug sudo paqet run -c config.yaml
```

## Contributing

We welcome contributions! See our contributing guide for details.

### Areas Needing Help

1. **KCP Implementation**: Port or optimize KCP protocol
2. **Performance**: Optimize hot paths with Cython
3. **Testing**: Increase test coverage
4. **Documentation**: Improve guides and examples
5. **Platform Support**: Test and fix Windows compatibility
6. **Features**: Add new transport protocols or features

## Resources

### Documentation
- [Python Migration Guide](docs/PYTHON_MIGRATION.md)
- [Project Structure](docs/PYTHON_PROJECT_STRUCTURE.md)
- [API Reference](docs/api/)

### External Resources
- [Original Go Implementation](https://github.com/amir0241/paqet)
- [KCP Protocol Specification](https://github.com/skywind3000/kcp)
- [QUIC RFC 9000](https://www.rfc-editor.org/rfc/rfc9000.html)
- [Scapy Documentation](https://scapy.readthedocs.io/)
- [aioquic Documentation](https://aioquic.readthedocs.io/)

### Community
- GitHub Issues: Report bugs and feature requests
- Discussions: Ask questions and share ideas
- Discord/Slack: Real-time chat (coming soon)

## Roadmap

### v0.1.0 (Milestone 1) - MVP
- ✅ Basic project structure
- ⏳ Configuration system
- ⏳ Raw socket handling
- ⏳ KCP transport
- ⏳ SOCKS5 proxy
- ⏳ Client mode
- ⏳ Server mode

### v0.2.0 (Milestone 2) - QUIC Support
- QUIC transport
- Certificate management
- Performance optimization

### v0.3.0 (Milestone 3) - Feature Parity
- All CLI commands
- Port forwarding
- Complete test coverage
- Comprehensive documentation

### v1.0.0 - Production Ready
- Stable API
- Optimized performance
- Cross-platform support
- PyPI release
- Binary distributions

## License

MIT License - Same as the original Go version

## Acknowledgments

- Original Go implementation by [amir0241](https://github.com/amir0241)
- Inspired by [gfw_resist_tcp_proxy](https://github.com/GFW-knocker/gfw_resist_tcp_proxy)
- Built with Python, Scapy, aioquic, and other open-source libraries

---

**Status**: 📝 Planning Phase  
**Next Milestone**: v0.1.0 MVP (Weeks 1-8)  
**Estimated Completion**: 12 weeks from start  
**Last Updated**: February 2026
