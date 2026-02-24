package conf

import (
	"runtime"
	"testing"
)

func TestSysRAMMB(t *testing.T) {
	got := sysRAMMB()
	if got <= 0 {
		t.Errorf("sysRAMMB() = %d, want > 0", got)
	}
	// Sanity: must be at least 64 MB (no modern machine has less).
	if got < 64 {
		t.Errorf("sysRAMMB() = %d MB, seems implausibly small", got)
	}
}

func TestSysCPUCount(t *testing.T) {
	got := sysCPUCount()
	want := runtime.NumCPU()
	if got != want {
		t.Errorf("sysCPUCount() = %d, want %d", got, want)
	}
	if got < 1 {
		t.Errorf("sysCPUCount() = %d, want >= 1", got)
	}
}

func TestClampInt(t *testing.T) {
	tests := []struct {
		v, lo, hi, want int
	}{
		{5, 1, 10, 5},     // within range
		{0, 1, 10, 1},     // below min
		{15, 1, 10, 10},   // above max
		{1, 1, 10, 1},     // at min
		{10, 1, 10, 10},   // at max
		{-5, -10, -1, -5}, // negative range
	}
	for _, tt := range tests {
		got := clampInt(tt.v, tt.lo, tt.hi)
		if got != tt.want {
			t.Errorf("clampInt(%d, %d, %d) = %d, want %d", tt.v, tt.lo, tt.hi, got, tt.want)
		}
	}
}

func TestNextPowerOf2(t *testing.T) {
	tests := []struct {
		v, want int
	}{
		{-1, 1},  // v <= 0: returns 1
		{0, 1},   // v <= 0: returns 1
		{1, 1},
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 8},
		{7, 8},
		{8, 8},
		{9, 16},
		{16, 16},
		{17, 32},
		{31, 32},
		{32, 32},
		{33, 64},
		{64, 64},
	}
	for _, tt := range tests {
		got := nextPowerOf2(tt.v)
		if got != tt.want {
			t.Errorf("nextPowerOf2(%d) = %d, want %d", tt.v, got, tt.want)
		}
	}
}

// TestKCPAutoTunedDefaults checks that KCP defaults are within the valid ranges
// produced by the auto-tuning formulas.
func TestKCPAutoTunedDefaults(t *testing.T) {
	for _, role := range []string{"client", "server"} {
		t.Run(role, func(t *testing.T) {
			k := &KCP{}
			k.setDefaults(role)

			// rcvwnd / sndwnd must be within KCP validate bounds.
			if k.Rcvwnd < 1 || k.Rcvwnd > 32768 {
				t.Errorf("Rcvwnd = %d, want in [1, 32768]", k.Rcvwnd)
			}
			if k.Sndwnd < 1 || k.Sndwnd > 32768 {
				t.Errorf("Sndwnd = %d, want in [1, 32768]", k.Sndwnd)
			}

			// smuxbuf >= 1024, streambuf >= 1024 (validate bounds).
			if k.Smuxbuf < 1024 {
				t.Errorf("Smuxbuf = %d, want >= 1024", k.Smuxbuf)
			}
			if k.Streambuf < 1024 {
				t.Errorf("Streambuf = %d, want >= 1024", k.Streambuf)
			}
		})
	}
}

// TestTransportAutoTunedDefaults checks that Transport buffer defaults are in valid ranges.
func TestTransportAutoTunedDefaults(t *testing.T) {
	for _, role := range []string{"client", "server"} {
		t.Run(role, func(t *testing.T) {
			tr := &Transport{Protocol: "kcp", KCP: &KCP{Block_: "none"}}
			tr.setDefaults(role)

			if tr.TCPBuf < 4*1024 || tr.TCPBuf > 16*1024*1024 {
				t.Errorf("TCPBuf = %d, want in [4KB, 16MB]", tr.TCPBuf)
			}
			if tr.UDPBuf < 2*1024 || tr.UDPBuf > 4*1024*1024 {
				t.Errorf("UDPBuf = %d, want in [2KB, 4MB]", tr.UDPBuf)
			}
			if tr.TUNBuf < 8*1024 || tr.TUNBuf > 32*1024*1024 {
				t.Errorf("TUNBuf = %d, want in [8KB, 32MB]", tr.TUNBuf)
			}
		})
	}
}

// TestPCAPAutoTunedDefaults checks that PCAP defaults are within valid validation bounds.
func TestPCAPAutoTunedDefaults(t *testing.T) {
	for _, role := range []string{"client", "server"} {
		t.Run(role, func(t *testing.T) {
			p := PCAP{}
			p.setDefaults(role)

			if p.Sockbuf < 1024 || p.Sockbuf > 100*1024*1024 {
				t.Errorf("Sockbuf = %d, want in [1KB, 100MB]", p.Sockbuf)
			}
			if p.SendQueueSize < 1 || p.SendQueueSize > 100000 {
				t.Errorf("SendQueueSize = %d, want in [1, 100000]", p.SendQueueSize)
			}

			// No validation errors should occur.
			if errs := p.validate(); len(errs) > 0 {
				t.Errorf("validate() returned errors: %v", errs)
			}
		})
	}
}

// TestPerformanceAutoTunedDefaults checks that Performance defaults are within valid validation bounds.
func TestPerformanceAutoTunedDefaults(t *testing.T) {
	for _, role := range []string{"client", "server"} {
		t.Run(role, func(t *testing.T) {
			p := Performance{}
			p.setDefaults(role)

			if p.PacketWorkers < 1 || p.PacketWorkers > 64 {
				t.Errorf("PacketWorkers = %d, want in [1, 64]", p.PacketWorkers)
			}
			if p.StreamWorkerPoolSize < 10 || p.StreamWorkerPoolSize > 100000 {
				t.Errorf("StreamWorkerPoolSize = %d, want in [10, 100000]", p.StreamWorkerPoolSize)
			}
			if p.TCPConnectionPoolSize < 0 || p.TCPConnectionPoolSize > 10000 {
				t.Errorf("TCPConnectionPoolSize = %d, want in [0, 10000]", p.TCPConnectionPoolSize)
			}

			// No validation errors should occur.
			if errs := p.validate(); len(errs) > 0 {
				t.Errorf("validate() returned errors: %v", errs)
			}
		})
	}
}

// TestAutoTunedCustomValuesPreserved checks that explicit values are not overridden.
func TestAutoTunedCustomValuesPreserved(t *testing.T) {
	k := &KCP{Rcvwnd: 256, Sndwnd: 256, Smuxbuf: 1024 * 1024, Streambuf: 512 * 1024}
	k.setDefaults("client")
	if k.Rcvwnd != 256 {
		t.Errorf("Rcvwnd was overridden: got %d, want 256", k.Rcvwnd)
	}
	if k.Sndwnd != 256 {
		t.Errorf("Sndwnd was overridden: got %d, want 256", k.Sndwnd)
	}
	if k.Smuxbuf != 1024*1024 {
		t.Errorf("Smuxbuf was overridden: got %d, want %d", k.Smuxbuf, 1024*1024)
	}
	if k.Streambuf != 512*1024 {
		t.Errorf("Streambuf was overridden: got %d, want %d", k.Streambuf, 512*1024)
	}

	p := PCAP{Sockbuf: 4 * 1024 * 1024, SendQueueSize: 1000}
	p.setDefaults("server")
	if p.Sockbuf != 4*1024*1024 {
		t.Errorf("Sockbuf was overridden: got %d, want %d", p.Sockbuf, 4*1024*1024)
	}
	if p.SendQueueSize != 1000 {
		t.Errorf("SendQueueSize was overridden: got %d, want 1000", p.SendQueueSize)
	}
}
