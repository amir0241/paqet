package conf

// GFWResist holds configuration options for GFW (Great Firewall) resistance
// using the TCP violation technique. This technique bypasses IP-based blocking
// by communicating via PSH+ACK packets, which the GFW does not inspect,
// instead of the standard TCP SYN handshake that the GFW checks against its
// blocklist.
//
// This concept is inspired by the gfw_resist_tcp_proxy project:
// https://github.com/GFW-knocker/gfw_resist_tcp_proxy
type GFWResist struct {
	// AutoIPTables controls whether paqet automatically applies and removes
	// the iptables rules required for TCP violation (GFW bypass) on the
	// server. When true, paqet adds NOTRACK rules to bypass kernel connection
	// tracking and a DROP rule to prevent the kernel from sending TCP RST
	// packets that would disrupt stateless raw-packet communication.
	//
	// Rules applied (for both iptables and ip6tables):
	//   iptables -t raw    -A PREROUTING -p tcp --dport PORT -j NOTRACK
	//   iptables -t raw    -A OUTPUT     -p tcp --sport PORT -j NOTRACK
	//   iptables -t mangle -A OUTPUT     -p tcp --sport PORT --tcp-flags RST RST -j DROP
	//
	// Rules are automatically removed when paqet shuts down. Requires root
	// privileges and is only supported on Linux.
	AutoIPTables bool `yaml:"auto_iptables"`
}

func (g *GFWResist) setDefaults() {}

func (g *GFWResist) validate() []error {
	return nil
}
