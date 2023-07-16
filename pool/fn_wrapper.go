package pool

import (
	"sync"
)

// Generator is a constructor function of some resource
type Generator[T any] func() T

// PoolWrapper is a generic wrapper over some function that requires a resource pool.
// Pooling is encapsulated in this function.
func PoolWrapper[T, In, Out any](generator Generator[T], fn func(T, In) Out) func(In) Out {
	pool := sync.Pool{New: func() interface{} {
		return generator()
	}}
	return func(in In) Out {
		t := pool.Get().(T)
		defer pool.Put(t)

		return fn(t, in)
	}
}

// PoolWrapperWithErr is a generic wrapper over some function that requires a resource pool,
// and returns error in addition to some output
func PoolWrapperWithErr[T, In, Out any](generator Generator[T], fn func(T, In) (Out, error)) func(In) (Out, error) {
	pool := sync.Pool{New: func() interface{} {
		return generator()
	}}
	return func(in In) (Out, error) {
		t := pool.Get().(T)
		defer pool.Put(t)

		return fn(t, in)
	}
}

// PoolWrapperErr is a generic wrapper over some function that requires a resource pool,
// the function is expected to return error only.
func PoolWrapperErr[T, In any](generator Generator[T], fn func(T, In) error) func(In) error {
	pool := sync.Pool{New: func() interface{} {
		return generator()
	}}
	return func(in In) error {
		t := pool.Get().(T)
		defer pool.Put(t)

		return fn(t, in)
	}
}
