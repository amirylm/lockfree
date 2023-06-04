package ringbuffer

import (
	"runtime"
	"sync/atomic"

	"github.com/amirylm/lockfree/common"
)

// NOTE: WIP

func New[V any](c int) *RingBuffer[V] {
	rb := &RingBuffer[V]{
		elements: make([]*atomic.Pointer[V], c),
		capacity: uint32(c),
	}

	for i := range rb.elements {
		rb.elements[i] = &atomic.Pointer[V]{}
	}

	return rb
}

// RingBuffer is a lock-free ring buffer implementation.
// NOTE: WIP
type RingBuffer[V any] struct {
	elements []*atomic.Pointer[V]

	capacity uint32
	state    atomic.Uint64
}

func (rb *RingBuffer[V]) Empty() bool {
	return newState(rb.state.Load()).IsEmpty()
}

func (rb *RingBuffer[V]) Full() bool {
	return newState(rb.state.Load()).full
}

// Push adds a new item to the buffer.
// In case of some conflict with other goroutine, we revert changes and call retry.
func (rb *RingBuffer[V]) Push(v V) error {
	originalState := rb.state.Load()
	state := newState(originalState)
	if state.full {
		return common.ErrOverflow
	}
	el := rb.elements[state.tail%rb.capacity]
	state.tail++
	state.full = (state.tail%rb.capacity == state.head%rb.capacity)
	orig := el.Swap(&v)
	if !rb.state.CompareAndSwap(originalState, state.Uint64()) {
		el.Store(orig)
		runtime.Gosched()
		return rb.Push(v)
	}
	return nil
}

// Enqueue pops the next item in the buffer.
// In case of some conflict with other goroutine, we revert changes and call retry.
func (rb *RingBuffer[V]) Pop() (V, bool) {
	originalState := rb.state.Load()
	state := newState(rb.state.Load())
	var v V
	if state.IsEmpty() {
		return v, false
	}
	el := rb.elements[state.head%rb.capacity]
	state.head++
	state.full = false
	val := el.Load()
	if !rb.state.CompareAndSwap(originalState, state.Uint64()) {
		// in case we have some conflict with another goroutine, retry.
		runtime.Gosched()
		return rb.Pop()
	}
	if val != nil {
		v = *val
	}
	return v, true
}

func (rb *RingBuffer[Value]) Size() int {
	state := newState(rb.state.Load())
	if state.full {
		return int(rb.capacity)
	}
	return int(state.tail - state.head)
}

// func (rb *RingBuffer[V]) Slice() []*V {
// 	res := make([]*V, 0)
// 	for _, el := range rb.elements {
// 		res = append(res, el.Load())
// 	}
// 	return res
// }
