package ringbuffer

import (
	"sync/atomic"

	"github.com/amirylm/lockfree/common"
)

// New creates a new RingBuffer
func New[V any](c int) common.DataStructure[V] {
	rb := &RingBuffer[V]{
		elements: make([]*atomic.Pointer[V], c),
		capacity: uint32(c),
		state:    atomic.Uint64{},
	}

	rb.state.Store(new(ringBufferState).Uint64())

	for i := range rb.elements {
		rb.elements[i] = &atomic.Pointer[V]{}
	}

	return rb
}

// RingBuffer is a lock-free ring buffer implementation.
type RingBuffer[V any] struct {
	elements []*atomic.Pointer[V]

	capacity uint32
	state    atomic.Uint64
}

func (rb *RingBuffer[V]) Empty() bool {
	return newState(rb.state.Load()).Empty()
}

func (rb *RingBuffer[V]) Full() bool {
	return newState(rb.state.Load()).Full()
}

func (rb *RingBuffer[Value]) Size() int {
	state := newState(rb.state.Load())
	if state.full {
		return int(rb.capacity)
	}
	return int(state.tail - state.head)
}

// Push adds a new item to the buffer.
// We revert changes and retry in case of some conflict with other goroutine.
func (rb *RingBuffer[V]) Push(v V) bool {
	originalState := rb.state.Load()
	state := newState(originalState)
	if state.full {
		return false
	}
	el := rb.elements[state.tail%rb.capacity]
	state.tail++
	state.full = state.tail%rb.capacity == state.head%rb.capacity
	if rb.state.CompareAndSwap(originalState, state.Uint64()) {
		el.Store(&v)
		return true
	}
	return rb.Push(v)
}

// Enqueue pops the next item in the buffer.
// We retry in case of some conflict with other goroutine.
func (rb *RingBuffer[V]) Pop() (V, bool) {
	originalState := rb.state.Load()
	state := newState(originalState)
	var v V
	if state.Empty() {
		return v, false
	}
	el := rb.elements[state.head%rb.capacity]
	state.head++
	state.full = false
	val := el.Load()
	if !rb.state.CompareAndSwap(originalState, state.Uint64()) {
		// in case we have some conflict with another goroutine, retry.
		return rb.Pop()
	}
	if val != nil {
		v = *val
	}
	return v, true
}
