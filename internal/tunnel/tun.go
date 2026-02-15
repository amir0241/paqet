package tunnel

import (
	"fmt"
	"net"
	"os/exec"
	"paqet/internal/conf"
	"paqet/internal/flog"
	"runtime"

	"github.com/songgao/water"
)

// TUN represents a TUN device for layer 3 networking
type TUN struct {
	iface *water.Interface
	cfg   *conf.TUN
}

// New creates and configures a new TUN device
func New(cfg *conf.TUN) (*TUN, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("TUN is not enabled in configuration")
	}

	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = cfg.Name

	iface, err := water.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN device: %v", err)
	}

	t := &TUN{
		iface: iface,
		cfg:   cfg,
	}

	if err := t.configure(); err != nil {
		iface.Close()
		return nil, err
	}

	flog.Infof("TUN device %s created with address %s", cfg.Name, cfg.Addr)
	return t, nil
}

// configure sets up the TUN interface with IP address and brings it up
func (t *TUN) configure() error {
	switch runtime.GOOS {
	case "linux":
		return t.configureLinux()
	case "darwin":
		return t.configureDarwin()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// configureLinux configures the TUN interface on Linux
func (t *TUN) configureLinux() error {
	// Set IP address
	cmd := exec.Command("ip", "addr", "add", t.cfg.Addr, "dev", t.cfg.Name)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set IP address: %v, output: %s", err, output)
	}

	// Set MTU
	cmd = exec.Command("ip", "link", "set", "dev", t.cfg.Name, "mtu", fmt.Sprintf("%d", t.cfg.MTU))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set MTU: %v, output: %s", err, output)
	}

	// Bring interface up
	cmd = exec.Command("ip", "link", "set", "dev", t.cfg.Name, "up")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to bring interface up: %v, output: %s", err, output)
	}

	return nil
}

// configureDarwin configures the TUN interface on macOS
func (t *TUN) configureDarwin() error {
	// Set IP address and destination (for point-to-point)
	// For macOS, we need to set both local and destination addresses
	ip := t.cfg.IP.String()
	network := t.cfg.Net
	
	// Calculate a destination address (typically the network address + 1 or last address - 1)
	destIP := make(net.IP, len(network.IP))
	copy(destIP, network.IP)
	destIP[len(destIP)-1]++
	if destIP.Equal(t.cfg.IP) {
		destIP[len(destIP)-1] += 1
	}

	cmd := exec.Command("ifconfig", t.cfg.Name, ip, destIP.String(), "up")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to configure interface: %v, output: %s", err, output)
	}

	// Set MTU
	cmd = exec.Command("ifconfig", t.cfg.Name, "mtu", fmt.Sprintf("%d", t.cfg.MTU))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set MTU: %v, output: %s", err, output)
	}

	return nil
}

// Read reads a packet from the TUN device.
// Note: TUN intentionally does NOT implement the io.ReaderFrom interface to ensure
// that io.CopyBuffer uses the provided 256KB buffer pool instead of allocating
// small MTU-sized buffers repeatedly, which significantly improves throughput.
func (t *TUN) Read(buf []byte) (int, error) {
	return t.iface.Read(buf)
}

// Write writes a packet to the TUN device.
// Note: TUN intentionally does NOT implement the io.WriterTo interface to ensure
// that io.CopyBuffer uses the provided 256KB buffer pool instead of allocating
// small MTU-sized buffers repeatedly, which significantly improves throughput.
func (t *TUN) Write(buf []byte) (int, error) {
	return t.iface.Write(buf)
}

// Close closes the TUN device
func (t *TUN) Close() error {
	return t.iface.Close()
}

// Name returns the interface name
func (t *TUN) Name() string {
	return t.cfg.Name
}
