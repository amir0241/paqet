package conf

import (
	"testing"
)

func TestTUNConfig(t *testing.T) {
	// Test TUN configuration parsing
	tun := TUN{
		Enabled: true,
		Name:    "tun0",
		Addr:    "10.0.8.1/24",
		MTU:     1400,
	}
	
	tun.setDefaults()
	
	if tun.Name != "tun0" {
		t.Errorf("Expected name 'tun0', got '%s'", tun.Name)
	}
	
	if tun.MTU != 1400 {
		t.Errorf("Expected MTU 1400, got %d", tun.MTU)
	}
	
	// Test validation
	errs := tun.validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors, got: %v", errs)
	}
	
	if tun.IP == nil {
		t.Error("Expected IP to be parsed")
	}
	
	if tun.Net == nil {
		t.Error("Expected Net to be parsed")
	}
	
	expectedIP := "10.0.8.1"
	if tun.IP.String() != expectedIP {
		t.Errorf("Expected IP %s, got %s", expectedIP, tun.IP.String())
	}
}

func TestTUNConfigInvalidAddr(t *testing.T) {
	tun := TUN{
		Enabled: true,
		Name:    "tun0",
		Addr:    "invalid",
		MTU:     1400,
	}
	
	errs := tun.validate()
	if len(errs) == 0 {
		t.Error("Expected validation error for invalid address")
	}
}

func TestTUNConfigDisabled(t *testing.T) {
	tun := TUN{
		Enabled: false,
	}
	
	errs := tun.validate()
	if len(errs) > 0 {
		t.Errorf("Expected no errors when TUN is disabled, got: %v", errs)
	}
}
