package conf

import (
	"fmt"
	"net"
)

type TUN struct {
	Enabled bool   `yaml:"enabled"`
	Name    string `yaml:"name"`
	Addr    string `yaml:"addr"`
	MTU     int    `yaml:"mtu"`

	IP  net.IP `yaml:"-"`
	Net *net.IPNet `yaml:"-"`
}

func (t *TUN) setDefaults() {
	if t.Name == "" {
		t.Name = "tun0"
	}
	if t.MTU == 0 {
		t.MTU = 1500
	}
}

func (t *TUN) validate() []error {
	var errors []error
	
	if !t.Enabled {
		return errors
	}

	if t.Addr == "" {
		errors = append(errors, fmt.Errorf("tun.addr is required when tun is enabled"))
		return errors
	}

	ip, ipNet, err := net.ParseCIDR(t.Addr)
	if err != nil {
		errors = append(errors, fmt.Errorf("invalid tun.addr format (expected CIDR, e.g., 10.0.0.1/24): %v", err))
		return errors
	}
	t.IP = ip
	t.Net = ipNet

	if t.MTU < 68 || t.MTU > 65535 {
		errors = append(errors, fmt.Errorf("tun.mtu must be between 68-65535"))
	}

	return errors
}
