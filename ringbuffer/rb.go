package ringbuffer

import (
	"sync/atomic"

	"github.com/amirylm/lockfree/common"
)

// New creates a new RingBuffer
func New[Value any](c int) common.DataStructure[Value] {
	rb := &RingBuffer[Value]{
		elements: make([]*atomic.Pointer[Value], c),
		capacity: uint32(c),
		state:    atomic.Uint64{},
	}

	rb.state.Store(new(ringBufferState).Uint64())

	for i := range rb.elements {
		rb.elements[i] = &atomic.Pointer[Value]{}
	}

	return rb
}

// RingBuffer is a lock-free ring buffer implementation.
type RingBuffer[Value any] struct {
	elements []*atomic.Pointer[Value]

	capacity uint32
	state    atomic.Uint64
}

func (rb *RingBuffer[Value]) Empty() bool {
	return newState(rb.state.Load()).Empty()
}

func (rb *RingBuffer[Value]) Full() bool {
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
func (rb *RingBuffer[Value]) Push(v Value) bool {
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

// Pop reads the next item in the buffer.
// We retry in case of some conflict with other goroutine.
func (rb *RingBuffer[Value]) Pop() (Value, bool) {
	originalState := rb.state.Load()
	state := newState(originalState)
	var v Value
	if state.Empty() {
		return v, false
	}
	el := rb.elements[state.head%rb.capacity]
	state.head++
	state.full = false
	if rb.state.CompareAndSwap(originalState, state.Uint64()) {
		val := el.Load()
		if val != nil {
			v = *val
		}
		return v, true
	}
	// in case we have some conflict with another goroutine, retry.
	return rb.Pop()
}
