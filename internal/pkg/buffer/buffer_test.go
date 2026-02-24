package buffer

import (
	"testing"
)

func TestInitialize(t *testing.T) {
	Initialize(4*1024, 2*1024, 8*1024)

	if TPool == nil || UPool == nil || TUNPool == nil {
		t.Fatal("Initialize must set all three pool variables")
	}

	if TPool.defaultSize != 4*1024 {
		t.Errorf("TPool.defaultSize = %d, want %d", TPool.defaultSize, 4*1024)
	}
	if UPool.defaultSize != 2*1024 {
		t.Errorf("UPool.defaultSize = %d, want %d", UPool.defaultSize, 2*1024)
	}
	if TUNPool.defaultSize != 8*1024 {
		t.Errorf("TUNPool.defaultSize = %d, want %d", TUNPool.defaultSize, 8*1024)
	}
}

func TestPoolGet(t *testing.T) {
	const size = 1024
	p := newPool(size)

	bufp := p.Get()
	if bufp == nil {
		t.Fatal("Get returned nil")
	}
	if len(*bufp) != size {
		t.Errorf("Get len = %d, want %d", len(*bufp), size)
	}
	if cap(*bufp) < size {
		t.Errorf("Get cap = %d, want >= %d", cap(*bufp), size)
	}
	p.Put(bufp)
}

func TestPoolGetN_WithinCapacity(t *testing.T) {
	const defaultSize = 1024
	p := newPool(defaultSize)

	// Request a smaller buffer — should be served from pool.
	small := 256
	bufp := p.GetN(small)
	if bufp == nil {
		t.Fatal("GetN returned nil")
	}
	if len(*bufp) != small {
		t.Errorf("GetN len = %d, want %d", len(*bufp), small)
	}
	if cap(*bufp) < defaultSize {
		t.Errorf("GetN cap = %d, want >= %d (pool-backed)", cap(*bufp), defaultSize)
	}
	p.Put(bufp)
}

func TestPoolGetN_BeyondCapacity(t *testing.T) {
	const defaultSize = 512
	p := newPool(defaultSize)

	// Request a larger buffer — must be a fresh allocation.
	large := 2 * 1024
	bufp := p.GetN(large)
	if bufp == nil {
		t.Fatal("GetN returned nil")
	}
	if len(*bufp) != large {
		t.Errorf("GetN len = %d, want %d", len(*bufp), large)
	}
	// The buffer was freshly allocated, so cap == large.
	if cap(*bufp) != large {
		t.Errorf("GetN cap = %d, want %d for fresh allocation", cap(*bufp), large)
	}
	// Putting an oversized buffer back must not pollute the pool.
	p.Put(bufp)

	// Next Get from pool should still return a properly-sized buffer.
	next := p.Get()
	if len(*next) != defaultSize {
		t.Errorf("after Put of oversized buf, Get len = %d, want %d", len(*next), defaultSize)
	}
	p.Put(next)
}

func TestPoolPutRestoresLength(t *testing.T) {
	const defaultSize = 1024
	p := newPool(defaultSize)

	bufp := p.GetN(128) // slice to 128
	p.Put(bufp)         // should restore length to defaultSize before returning to pool

	bufp2 := p.Get()
	if len(*bufp2) != defaultSize {
		t.Errorf("after Put, Get len = %d, want %d", len(*bufp2), defaultSize)
	}
	p.Put(bufp2)
}

func TestPoolGetN_ExactCapacity(t *testing.T) {
	const defaultSize = 1024
	p := newPool(defaultSize)

	// Request exactly the default size.
	bufp := p.GetN(defaultSize)
	if len(*bufp) != defaultSize {
		t.Errorf("GetN(defaultSize) len = %d, want %d", len(*bufp), defaultSize)
	}
	p.Put(bufp)
}

func TestPoolReuseAfterPut(t *testing.T) {
	const defaultSize = 1024
	p := newPool(defaultSize)

	bufp1 := p.Get()
	ptr1 := &(*bufp1)[0]
	p.Put(bufp1)

	bufp2 := p.Get()
	ptr2 := &(*bufp2)[0]
	p.Put(bufp2)

	// sync.Pool may or may not reuse the same backing array, but if it does
	// the pointers should match. We just check we don't panic.
	_ = ptr1
	_ = ptr2
}
