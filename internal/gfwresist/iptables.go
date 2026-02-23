package gfwresist

import (
	"fmt"
	"os/exec"
	"runtime"

	"paqet/internal/flog"
)

// iptablesRule describes a single iptables rule.
type iptablesRule struct {
	table string
	chain string
	args  []string
}

// appliedRule records a rule that was successfully added so it can be
// precisely removed during cleanup even if Apply() only partially succeeded.
type appliedRule struct {
	bin  string
	rule iptablesRule
}

// IPTablesManager manages the iptables rules required for TCP violation
// (GFW bypass) operation on Linux. It applies NOTRACK rules to bypass
// connection tracking and drops outbound RST packets that the kernel
// would otherwise send in response to stateless raw TCP packets.
type IPTablesManager struct {
	port    int
	applied []appliedRule
}

// NewIPTablesManager creates a manager for the given server port.
func NewIPTablesManager(port int) *IPTablesManager {
	return &IPTablesManager{port: port}
}

// Apply adds the required iptables (and ip6tables) rules for the server port.
// Returns an error if the platform is not Linux, the port is invalid, or if
// any rule fails to apply. Successfully applied rules are tracked so that
// Cleanup can remove them precisely even after a partial failure.
func (m *IPTablesManager) Apply() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("auto_iptables is only supported on Linux (current OS: %s)", runtime.GOOS)
	}
	if m.port < 1 || m.port > 65535 {
		return fmt.Errorf("invalid port %d: must be between 1 and 65535", m.port)
	}

	rules := m.rules()
	for _, rule := range rules {
		for _, bin := range []string{"iptables", "ip6tables"} {
			args := append([]string{"-t", rule.table, "-A", rule.chain}, rule.args...)
			if err := runCmd(bin, args...); err != nil {
				return fmt.Errorf("failed to add %s rule (table=%s chain=%s): %w", bin, rule.table, rule.chain, err)
			}
			m.applied = append(m.applied, appliedRule{bin: bin, rule: rule})
			flog.Debugf("applied %s rule: -t %s -A %s %v", bin, rule.table, rule.chain, rule.args)
		}
	}

	flog.Infof("GFW-resist: iptables rules applied for port %d", m.port)
	return nil
}

// Cleanup removes only the iptables rules that were successfully applied by
// Apply. Errors are logged but not returned to ensure cleanup always completes.
func (m *IPTablesManager) Cleanup() {
	for _, a := range m.applied {
		args := append([]string{"-t", a.rule.table, "-D", a.rule.chain}, a.rule.args...)
		if err := runCmd(a.bin, args...); err != nil {
			flog.Warnf("failed to remove %s rule (table=%s chain=%s): %v", a.bin, a.rule.table, a.rule.chain, err)
		} else {
			flog.Debugf("removed %s rule: -t %s -D %s %v", a.bin, a.rule.table, a.rule.chain, a.rule.args)
		}
	}
	if len(m.applied) > 0 {
		flog.Infof("GFW-resist: iptables rules cleaned up for port %d", m.port)
	}
}

// rules returns the set of iptables rules required for TCP violation operation.
//
// Rule 1 & 2: Bypass kernel connection tracking (conntrack) for the server port.
//   - Without NOTRACK, the kernel tracks these stateless raw TCP packets as
//     "INVALID" connections and may drop them or send RST.
//
// Rule 3: Prevent the kernel from sending TCP RST packets from the server port.
//   - When the kernel receives a PSH+ACK packet with no matching connection,
//     it generates a RST response. This RST can break stateful NAT/firewall
//     state on intermediate devices. Dropping it keeps the channel open.
func (m *IPTablesManager) rules() []iptablesRule {
	port := fmt.Sprintf("%d", m.port)
	return []iptablesRule{
		{"raw", "PREROUTING", []string{"-p", "tcp", "--dport", port, "-j", "NOTRACK"}},
		{"raw", "OUTPUT", []string{"-p", "tcp", "--sport", port, "-j", "NOTRACK"}},
		{"mangle", "OUTPUT", []string{"-p", "tcp", "--sport", port, "--tcp-flags", "RST", "RST", "-j", "DROP"}},
	}
}

func runCmd(name string, args ...string) error {
	path, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("%s not found in PATH: %w", name, err)
	}
	out, err := exec.Command(path, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(out))
	}
	return nil
}
