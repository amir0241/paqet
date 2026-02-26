package conf

import (
	"fmt"
	"slices"
)

type Transport struct {
	Protocol string `yaml:"protocol"`
	Conn     int    `yaml:"conn"`
	TCPBuf   int    `yaml:"tcpbuf"`
	UDPBuf   int    `yaml:"udpbuf"`
	TUNBuf   int    `yaml:"tunbuf"`
	KCP      *KCP   `yaml:"kcp"`
	QUIC     *QUIC  `yaml:"quic"`
}

func (t *Transport) setDefaults(role string) {
	cpus := sysCPUCount()
	if t.Protocol == "" {
		t.Protocol = "quic"
	}

	if t.Conn == 0 {
		if role == "client" {
			switch t.Protocol {
			case "quic":
				t.Conn = clampInt(cpus/2, 1, 4)
			case "kcp":
				t.Conn = clampInt(cpus/3, 1, 3)
			default:
				t.Conn = 1
			}
		} else {
			t.Conn = 1
		}
	}

	if t.TCPBuf == 0 {
		// Scale with CPU count: 16 KB per core, between 64 KB and 4 MB.
		t.TCPBuf = clampInt(cpus*16*1024, 64*1024, 4*1024*1024)
	}
	if t.TCPBuf < 4*1024 {
		t.TCPBuf = 4 * 1024
	}
	if t.UDPBuf == 0 {
		// Scale with CPU count: 4 KB per core, between 16 KB and 1 MB.
		t.UDPBuf = clampInt(cpus*4*1024, 16*1024, 1*1024*1024)
	}
	if t.UDPBuf < 2*1024 {
		t.UDPBuf = 2 * 1024
	}
	if t.TUNBuf == 0 {
		// Scale with CPU count: 64 KB per core, between 256 KB and 16 MB.
		t.TUNBuf = clampInt(cpus*64*1024, 256*1024, 16*1024*1024)
	}
	if t.TUNBuf < 8*1024 {
		t.TUNBuf = 8 * 1024
	}

	switch t.Protocol {
	case "kcp":
		if t.KCP == nil {
			t.KCP = &KCP{}
		}
		t.KCP.setDefaults(role)
	case "quic":
		if t.QUIC == nil {
			t.QUIC = &QUIC{}
		}
		t.QUIC.setDefaults(role)
	}
}

func (t *Transport) validate() []error {
	var errors []error

	validProtocols := []string{"kcp", "quic"}
	if !slices.Contains(validProtocols, t.Protocol) {
		errors = append(errors, fmt.Errorf("transport protocol must be one of: %v", validProtocols))
	}

	if t.Conn < 1 || t.Conn > 256 {
		errors = append(errors, fmt.Errorf("KCP conn must be between 1-256 connections"))
	}

	if t.TCPBuf < 4*1024 || t.TCPBuf > 16*1024*1024 {
		errors = append(errors, fmt.Errorf("tcpbuf must be between 4KB and 16MB"))
	}

	if t.UDPBuf < 2*1024 || t.UDPBuf > 4*1024*1024 {
		errors = append(errors, fmt.Errorf("udpbuf must be between 2KB and 4MB"))
	}

	if t.TUNBuf < 8*1024 || t.TUNBuf > 32*1024*1024 {
		errors = append(errors, fmt.Errorf("tunbuf must be between 8KB and 32MB"))
	}

	switch t.Protocol {
	case "kcp":
		if t.KCP == nil {
			errors = append(errors, fmt.Errorf("transport.kcp is required when protocol is 'kcp'"))
			return errors
		}
		errors = append(errors, t.KCP.validate()...)
	case "quic":
		if t.QUIC == nil {
			errors = append(errors, fmt.Errorf("transport.quic is required when protocol is 'quic'"))
			return errors
		}
		errors = append(errors, t.QUIC.validate()...)
	}

	return errors
}
