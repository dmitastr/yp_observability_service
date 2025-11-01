package pool

import "sync"

// Resetter defines a type that can reset its internal state.
type Resetter interface {
	Reset()
}

// Pool provides a generic object pool for types implementing the Resetter interface.
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New creates and returns a new Pool instance for the given type.
func New[T Resetter]() *Pool[T] {
	return &Pool[T]{pool: sync.Pool{}}
}

// Put resets the given object and returns it to the pool for reuse.
func (p *Pool[T]) Put(x T) {
	x.Reset()
	p.pool.Put(x)
}

// Get retrieves an object from the pool and returns it as a pointer to T.
// The caller must ensure the object is properly initialized before use.
func (p *Pool[T]) Get() *T {
	return p.pool.Get().(*T)
}
