package buffer

import (
	"fmt"
	"sync"
)

var (
	TPool sync.Pool
	UPool sync.Pool
)

const (
	// Sensible limits for buffer sizes
	MinBufferSize = 1024        // 1KB minimum
	MaxBufferSize = 10 * 1024 * 1024 // 10MB maximum to prevent excessive memory allocation
	DefaultTCPBufferSize = 32 * 1024 // 32KB for TCP
	DefaultUDPBufferSize = 64 * 1024 // 64KB for UDP
)

func Initialize(tPool, uPool int) error {
	// Validate TCP buffer size
	if tPool < MinBufferSize || tPool > MaxBufferSize {
		return fmt.Errorf("invalid TCP buffer size %d, must be between %d and %d", tPool, MinBufferSize, MaxBufferSize)
	}
	
	// Validate UDP buffer size
	if uPool < MinBufferSize || uPool > MaxBufferSize {
		return fmt.Errorf("invalid UDP buffer size %d, must be between %d and %d", uPool, MinBufferSize, MaxBufferSize)
	}
	
	TPool = sync.Pool{
		New: func() any {
			b := make([]byte, tPool)
			return &b
		},
	}
	UPool = sync.Pool{
		New: func() any {
			b := make([]byte, uPool)
			return &b
		},
	}
	
	return nil
}
