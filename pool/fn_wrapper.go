package pool

import (
	"sync"
)

// Generator is a constructor function of some resource
type Generator[T any] func() T

// PoolWrapper is a generic wrapper over some function that requires a resource pool.
// Under the hood, sync.Pool is used to facilitate pooling.
//
// Usage example:
//
//	 	hashFn := PoolWrapper(sha256.New, func(h hash.Hash, data []byte) [32]byte {
//			var b [32]byte
//			_, err := h.Write(data)
//			defer h.Reset()
//			if err != nil {
//				return b
//			}
//			h.Sum(b[:0])
//			return b
//		})
//		h := hashFn([]byte("dummy bytes"))
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
// and returns error in addition to some output.
// Under the hood, sync.Pool is used to facilitate pooling.
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
// Under the hood, sync.Pool is used to facilitate pooling.
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
