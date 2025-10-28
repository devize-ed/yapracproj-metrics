package pool

import (
	"sync"
)

// Reseter is an interface that defines the Reset method.
type Reseter interface {
	Reset()
}

// Pool is a pool of objects that implement the Reseter interface.
type Pool[T Reseter] struct {
	pool *sync.Pool
}

// NewPool creates a new pool of objects that implement the Reseter interface.
func NewPool[T Reseter](new func() T) *Pool[T] {
	pool := &sync.Pool{
		New: func() any {
			return new()
		},
	}
	return &Pool[T]{
		pool: pool,
	}
}

// Get retrieves an object from the pool.
func (p *Pool[T]) Get() T {
	obj := p.pool.Get()
	// Assert the object is of type T and return
	return obj.(T)
}

// Put returns an object to the pool.
func (p *Pool[T]) Put(x T) {
	// Reset the object.
	x.Reset()
	// Put the object back into the pool.
	p.pool.Put(x)
}
