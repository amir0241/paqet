package buffer

import (
	"sync"
)

// Pool wraps sync.Pool with a fixed default buffer size and supports dynamic-size requests.
type Pool struct {
	pool        sync.Pool
	defaultSize int
}

// newPool creates a Pool whose New function allocates buffers of size bytes.
func newPool(size int) *Pool {
	p := &Pool{defaultSize: size}
	p.pool.New = func() any {
		b := make([]byte, size)
		return &b
	}
	return p
}

// Get returns a *[]byte of the pool's default size.
func (p *Pool) Get() *[]byte {
	bufp := p.pool.Get().(*[]byte)
	*bufp = (*bufp)[:p.defaultSize]
	return bufp
}

// GetN returns a *[]byte of exactly n bytes.
// If n is within the pool's default capacity the underlying pool buffer is reused;
// otherwise a fresh allocation of size n is returned (and Put is a no-op for it).
func (p *Pool) GetN(n int) *[]byte {
	bufp := p.pool.Get().(*[]byte)
	if cap(*bufp) >= n {
		*bufp = (*bufp)[:n]
		return bufp
	}
	// Pool buffer too small; return it and allocate exactly what is needed.
	p.pool.Put(bufp)
	b := make([]byte, n)
	return &b
}

// Put returns bufp to the pool.
// Buffers whose capacity is smaller than the pool's default size are discarded
// so they do not pollute the pool with undersized entries.
func (p *Pool) Put(bufp *[]byte) {
	if cap(*bufp) < p.defaultSize {
		return
	}
	*bufp = (*bufp)[:p.defaultSize]
	p.pool.Put(bufp)
}

var (
	TPool   *Pool
	UPool   *Pool
	TUNPool *Pool
)

func Initialize(tPool, uPool, tunPool int) {
	TPool = newPool(tPool)
	UPool = newPool(uPool)
	TUNPool = newPool(tunPool)
}
