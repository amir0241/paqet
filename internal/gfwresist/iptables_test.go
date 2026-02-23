package gfwresist

import (
	"runtime"
	"slices"
	"testing"
)

// TestNewIPTablesManager verifies the manager is created with the correct port.
func TestNewIPTablesManager(t *testing.T) {
	mgr := NewIPTablesManager(9999)
	if mgr == nil {
		t.Fatal("expected non-nil IPTablesManager")
	}
	if mgr.port != 9999 {
		t.Errorf("expected port 9999, got %d", mgr.port)
	}
}

// TestIPTablesManagerRules verifies the correct iptables rules are generated.
func TestIPTablesManagerRules(t *testing.T) {
	mgr := NewIPTablesManager(8888)
	rules := mgr.rules()

	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}

	// Rule 1: raw PREROUTING NOTRACK for incoming packets
	r := rules[0]
	if r.table != "raw" || r.chain != "PREROUTING" {
		t.Errorf("rule 0: expected raw/PREROUTING, got %s/%s", r.table, r.chain)
	}
	if !slices.Contains(r.args, "NOTRACK") {
		t.Errorf("rule 0: expected NOTRACK, got %v", r.args)
	}
	if !slices.Contains(r.args, "--dport") || !slices.Contains(r.args, "8888") {
		t.Errorf("rule 0: expected --dport 8888, got %v", r.args)
	}

	// Rule 2: raw OUTPUT NOTRACK for outgoing packets
	r = rules[1]
	if r.table != "raw" || r.chain != "OUTPUT" {
		t.Errorf("rule 1: expected raw/OUTPUT, got %s/%s", r.table, r.chain)
	}
	if !slices.Contains(r.args, "NOTRACK") {
		t.Errorf("rule 1: expected NOTRACK, got %v", r.args)
	}
	if !slices.Contains(r.args, "--sport") || !slices.Contains(r.args, "8888") {
		t.Errorf("rule 1: expected --sport 8888, got %v", r.args)
	}

	// Rule 3: mangle OUTPUT DROP RST
	r = rules[2]
	if r.table != "mangle" || r.chain != "OUTPUT" {
		t.Errorf("rule 2: expected mangle/OUTPUT, got %s/%s", r.table, r.chain)
	}
	if !slices.Contains(r.args, "DROP") {
		t.Errorf("rule 2: expected DROP, got %v", r.args)
	}
	if !slices.Contains(r.args, "RST") {
		t.Errorf("rule 2: expected RST flag, got %v", r.args)
	}
}

// TestApplyOnNonLinux verifies that Apply returns an error on non-Linux platforms.
func TestApplyOnNonLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("skipping non-Linux test on Linux")
	}
	mgr := NewIPTablesManager(9999)
	err := mgr.Apply()
	if err == nil {
		t.Error("expected error on non-Linux platform, got nil")
	}
}

// TestCleanupOnNonLinux verifies that Cleanup is a no-op (no panic) on non-Linux.
func TestCleanupOnNonLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("skipping non-Linux test on Linux")
	}
	mgr := NewIPTablesManager(9999)
	// Should not panic
	mgr.Cleanup()
}

// TestApplyInvalidPort verifies that Apply returns an error for invalid ports.
func TestApplyInvalidPort(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("port validation only runs after OS check, skip on non-Linux")
	}
	for _, port := range []int{0, -1, 65536, 99999} {
		mgr := NewIPTablesManager(port)
		err := mgr.Apply()
		if err == nil {
			t.Errorf("expected error for port %d, got nil", port)
		}
	}
}

// TestCleanupTracksAppliedRules verifies that Cleanup only removes rules that
// were successfully applied (applied slice is empty before Apply).
func TestCleanupTracksAppliedRules(t *testing.T) {
	mgr := NewIPTablesManager(9999)
	// No Apply called: applied slice should be empty
	if len(mgr.applied) != 0 {
		t.Errorf("expected 0 applied rules before Apply(), got %d", len(mgr.applied))
	}
	// Cleanup with empty applied list should be a no-op (no panic)
	mgr.Cleanup()
}
