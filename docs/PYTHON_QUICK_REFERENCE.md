# Quick Reference Guide: Python Port of paqet

This is a quick reference for developers working on porting paqet to Python.

## Repository Links

- **Original Go Repository**: https://github.com/amir0241/paqet
- **Python Version** (to be created): Create a new repository for the Python version

## Quick Start Commands

### Setting Up New Python Repository

```bash
# 1. Create new repository
mkdir paqet-python && cd paqet-python
git init

# 2. Set up Python environment
python3 -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate

# 3. Install development tools
pip install --upgrade pip
pip install black isort mypy pytest pytest-asyncio

# 4. Create basic structure
mkdir -p paqet/{cli,config,client,server,transport,socket,socks5,protocol,crypto,utils}
mkdir -p tests/{unit,integration} examples docs

# 5. Create __init__ files
find paqet -type d -exec touch {}/__init__.py \;
```

### Project Files to Create First

1. **paqet/__init__.py** - Package initialization with version
2. **setup.py** - Package setup configuration
3. **pyproject.toml** - Modern Python packaging
4. **requirements.txt** - Production dependencies
5. **requirements-dev.txt** - Development dependencies
6. **.gitignore** - Python gitignore
7. **README.md** - Project documentation
8. **LICENSE** - MIT License (copy from Go version)

## Module Implementation Order

### Priority 1: Foundation (Start Here)

1. **paqet/config/models.py** - Configuration data classes
2. **paqet/config/loader.py** - YAML config loader  
3. **paqet/logging/logger.py** - Logging setup
4. **paqet/cli/main.py** - CLI entry point

### Priority 2: Network Layer

5. **paqet/socket/raw_socket.py** - Raw socket with pcap
6. **paqet/socket/packet_crafter.py** - TCP packet crafting
7. **paqet/socket/buffer.py** - Buffer management
8. **paqet/utils/checksum.py** - Checksum calculation

### Priority 3: Transport

9. **paqet/transport/base.py** - Abstract transport interface
10. **paqet/transport/kcp/protocol.py** - KCP implementation (most complex)
11. **paqet/transport/kcp/connection.py** - KCP connection
12. **paqet/transport/quic/connection.py** - QUIC wrapper

### Priority 4: Application Layer

13. **paqet/socks5/server.py** - SOCKS5 server
14. **paqet/client/client.py** - Client mode
15. **paqet/server/server.py** - Server mode
16. **paqet/cli/run.py** - Run command

### Priority 5: Utilities

17. **paqet/cli/ping.py** - Ping command
18. **paqet/cli/dump.py** - Dump command
19. **paqet/cli/secret.py** - Secret generation
20. **paqet/cli/iface.py** - Interface listing

## Key Python Libraries

### Installation Commands

```bash
# Core networking
pip install scapy pypcap

# QUIC support
pip install aioquic

# CLI framework
pip install click

# Configuration
pip install PyYAML pydantic

# Cryptography
pip install cryptography

# Development
pip install pytest pytest-asyncio black mypy
```

### Library Usage Examples

#### Scapy - Packet Crafting

```python
from scapy.all import Ether, IP, TCP, Raw

# Craft a TCP packet
packet = Ether(src="aa:bb:cc:dd:ee:ff", dst="11:22:33:44:55:66") / \
         IP(src="192.168.1.100", dst="10.0.0.100") / \
         TCP(sport=12345, dport=9999, flags="PA") / \
         Raw(load=b"Hello")

# Get raw bytes
raw_bytes = bytes(packet)

# Parse packet
parsed = Ether(raw_bytes)
print(parsed.summary())
```

#### Pypcap - Raw Socket

```python
import pcap

# Open interface
pc = pcap.pcap(name="eth0", snaplen=65535, promisc=True)

# Send packet
pc.inject(raw_packet_bytes)

# Receive packets
for timestamp, packet in pc.readpkts():
    print(f"Captured {len(packet)} bytes")
```

#### aioquic - QUIC

```python
from aioquic.asyncio import connect
from aioquic.quic.configuration import QuicConfiguration

# Create config
config = QuicConfiguration(is_client=True)

# Connect
async with connect("example.com", 4433, configuration=config) as protocol:
    # Use QUIC connection
    pass
```

#### Click - CLI

```python
import click

@click.group()
def cli():
    """Main CLI"""
    pass

@cli.command()
@click.option('-c', '--config', required=True)
def run(config):
    """Run paqet"""
    print(f"Running with config: {config}")

if __name__ == '__main__':
    cli()
```

#### Pydantic - Config Validation

```python
from pydantic import BaseModel, Field, validator

class Config(BaseModel):
    role: str
    port: int = Field(ge=1, le=65535)
    
    @validator('role')
    def validate_role(cls, v):
        if v not in ['client', 'server']:
            raise ValueError('Invalid role')
        return v

# Use it
config = Config(role='client', port=1080)
```

## Code Patterns to Follow

### Async/Await Pattern

```python
import asyncio

async def handle_connection(reader, writer):
    """Handle incoming connection"""
    try:
        data = await reader.read(1024)
        # Process data
        writer.write(response)
        await writer.drain()
    finally:
        writer.close()
        await writer.wait_closed()

async def main():
    server = await asyncio.start_server(
        handle_connection, '127.0.0.1', 8888)
    async with server:
        await server.serve_forever()

# Run
asyncio.run(main())
```

### Error Handling

```python
import logging

logger = logging.getLogger(__name__)

def risky_operation():
    try:
        # Do something
        result = operation()
        return result
    except SpecificError as e:
        logger.error(f"Specific error: {e}")
        raise
    except Exception as e:
        logger.exception(f"Unexpected error: {e}")
        raise RuntimeError("Operation failed") from e
```

### Type Hints

```python
from typing import Optional, List, Dict, Union
from pathlib import Path

def process_data(
    data: bytes,
    config: Dict[str, Union[str, int]],
    output: Optional[Path] = None
) -> List[str]:
    """
    Process data according to config.
    
    Args:
        data: Input data bytes
        config: Configuration dictionary
        output: Optional output file path
        
    Returns:
        List of processed strings
        
    Raises:
        ValueError: If data is invalid
    """
    pass
```

## Testing Patterns

### Unit Test Example

```python
import pytest
from paqet.config.loader import load_config

def test_load_config_success(tmp_path):
    # Create test config file
    config_file = tmp_path / "config.yaml"
    config_file.write_text("""
    role: client
    network:
      interface: eth0
    """)
    
    # Load and validate
    config = load_config(config_file)
    assert config.role == "client"
    assert config.network.interface == "eth0"

def test_load_config_invalid():
    with pytest.raises(ValueError):
        load_config("nonexistent.yaml")
```

### Async Test Example

```python
import pytest

@pytest.mark.asyncio
async def test_connection():
    from paqet.client import PaqetClient
    
    client = PaqetClient(config)
    await client.connect()
    assert client.is_connected()
    await client.disconnect()
```

### Mock Example

```python
from unittest.mock import Mock, patch

def test_with_mock():
    with patch('paqet.socket.RawSocket') as mock_socket:
        mock_socket.return_value.send_packet.return_value = 100
        
        # Use mocked socket
        result = function_using_socket()
        assert result == expected
        mock_socket.send_packet.assert_called_once()
```

## Common Patterns from Go to Python

### Goroutines → asyncio

```python
# Go:
# go func() { doWork() }()

# Python:
import asyncio
asyncio.create_task(do_work())
```

### Channels → Queues

```python
# Go:
# ch := make(chan string)
# ch <- "message"
# msg := <-ch

# Python:
import asyncio
queue = asyncio.Queue()
await queue.put("message")
msg = await queue.get()
```

### Select → wait/gather

```python
# Go:
# select {
#   case <-ch1:
#   case <-ch2:
# }

# Python:
import asyncio
done, pending = await asyncio.wait(
    [task1, task2],
    return_when=asyncio.FIRST_COMPLETED
)
```

### Defer → try/finally

```python
# Go:
# defer cleanup()

# Python:
try:
    do_work()
finally:
    cleanup()

# Or use context manager
with resource:
    do_work()  # cleanup automatic
```

## Performance Optimization Tips

### 1. Use Cython for Hot Paths

```python
# kcp_fast.pyx (Cython)
cdef class FastKCP:
    cdef unsigned int conv
    cdef unsigned int mtu
    
    cpdef send_packet(self, bytes data):
        # Fast C-level code
        pass
```

### 2. Profile Before Optimizing

```bash
# CPU profiling
python -m cProfile -o output.prof script.py
python -m pstats output.prof

# Memory profiling
pip install memory_profiler
python -m memory_profiler script.py
```

### 3. Use PyPy for Production

```bash
# Install PyPy
pypy3 -m pip install paqet

# Run with PyPy (often 2-5x faster)
sudo pypy3 -m paqet run -c config.yaml
```

### 4. Optimize Data Structures

```python
# Use __slots__ for memory efficiency
class Packet:
    __slots__ = ['data', 'timestamp', 'size']
    
    def __init__(self, data, timestamp):
        self.data = data
        self.timestamp = timestamp
        self.size = len(data)
```

## Debugging Tips

### 1. Packet Debugging with Scapy

```python
from scapy.all import *

# Capture and display packets
sniff(iface="eth0", prn=lambda x: x.show(), count=10)

# Hexdump
hexdump(packet)

# Interactive exploration
packet.show2()  # Show computed fields
```

### 2. Async Debugging

```python
import asyncio
import logging

# Enable asyncio debug mode
asyncio.run(main(), debug=True)

# Log slow callbacks
logging.basicConfig(level=logging.DEBUG)
asyncio.get_event_loop().slow_callback_duration = 0.1
```

### 3. Network Debugging

```bash
# Monitor packets
sudo tcpdump -i eth0 -n port 9999 -X

# Check interface stats
ip -s link show eth0

# Test raw socket permissions
python3 -c "import socket; socket.socket(socket.AF_PACKET, socket.SOCK_RAW)"
```

## Documentation Standards

### Docstring Format

```python
def function(arg1: str, arg2: int = 0) -> bool:
    """
    Short one-line description.
    
    Longer description if needed. Explain what the function
    does in detail.
    
    Args:
        arg1: Description of arg1
        arg2: Description of arg2. Defaults to 0.
        
    Returns:
        True if successful, False otherwise.
        
    Raises:
        ValueError: If arg1 is empty
        RuntimeError: If operation fails
        
    Example:
        >>> function("test", 5)
        True
    """
    pass
```

## Git Workflow

### Branching

```bash
# Feature branch
git checkout -b feature/kcp-implementation

# Work and commit
git add .
git commit -m "feat: implement KCP connection handling"

# Push and create PR
git push origin feature/kcp-implementation
```

### Commit Messages

```
feat: add new feature
fix: fix bug
docs: update documentation
test: add tests
refactor: refactor code
perf: improve performance
chore: maintenance tasks
```

## CI/CD Setup

### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        python-version: [3.8, 3.9, '3.10', '3.11']
    
    steps:
    - uses: actions/checkout@v2
    - name: Set up Python
      uses: actions/setup-python@v2
      with:
        python-version: ${{ matrix.python-version }}
    
    - name: Install system dependencies
      run: |
        sudo apt-get update
        sudo apt-get install -y libpcap-dev
    
    - name: Install Python dependencies
      run: |
        pip install -e ".[dev]"
    
    - name: Run tests
      run: |
        pytest --cov=paqet
    
    - name: Lint
      run: |
        black --check paqet tests
        mypy paqet
```

## Resources

### Essential Reading
- [Scapy Tutorial](https://scapy.readthedocs.io/en/latest/usage.html)
- [Python asyncio](https://docs.python.org/3/library/asyncio.html)
- [Click Documentation](https://click.palletsprojects.com/)
- [Pydantic Documentation](https://docs.pydantic.dev/)
- [KCP Protocol](https://github.com/skywind3000/kcp)

### Tools
- [PyCharm](https://www.jetbrains.com/pycharm/) - IDE
- [VS Code](https://code.visualstudio.com/) + Python extension
- [ipython](https://ipython.org/) - Enhanced Python shell
- [Wireshark](https://www.wireshark.org/) - Packet analysis

## Getting Help

1. **Check Documentation**: See docs/ directory
2. **Review Go Implementation**: Compare with original code
3. **Ask Questions**: Create GitHub issue/discussion
4. **Join Community**: (Discord/Slack links TBD)

## Next Steps

1. ✅ Read this guide
2. ⏳ Set up development environment
3. ⏳ Start with Priority 1 modules
4. ⏳ Write tests as you go
5. ⏳ Review and iterate

---

**Good luck with the Python port! 🐍**
