package ringbuffer

import (
	"sync/atomic"

	"github.com/amirylm/go-options"
	"github.com/amirylm/lockfree/core"
)

func WithCapacity[Value any](c int) options.Option[RingBuffer[Value]] {
	return func(rb *RingBuffer[Value]) {
		rb.capacity = uint32(c)
	}
}

func WithOverride[Value any]() options.Option[RingBuffer[Value]] {
	return func(rb *RingBuffer[Value]) {
		rb.override = true
	}
}

// New creates a new RingBuffer
func New[Value any](opts ...options.Option[RingBuffer[Value]]) core.Queue[Value] {
	rb := &RingBuffer[Value]{
		state: atomic.Uint64{},
	}

	_ = options.Apply(rb, opts...)
	rb.elements = make([]*atomic.Pointer[Value], rb.capacity)

	rb.state.Store(new(ringBufferState).Uint64())

	for i := range rb.elements {
		rb.elements[i] = &atomic.Pointer[Value]{}
	}

	return rb
}

// RingBuffer is a lock-free queue implementation based on a ring buffer.
type RingBuffer[Value any] struct {
	elements []*atomic.Pointer[Value]

	capacity uint32
	state    atomic.Uint64

	override bool
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
func (rb *RingBuffer[Value]) Enqueue(v Value) bool {
	originalState := rb.state.Load()
	state := newState(originalState)
	if state.full {
		if !rb.override {
			return false
		}
		// in case we override items
		_, _ = rb.Dequeue()
	}
	el := rb.elements[state.tail%rb.capacity]
	state.tail++
	state.full = state.tail%rb.capacity == state.head%rb.capacity
	if rb.state.CompareAndSwap(originalState, state.Uint64()) {
		el.Store(&v)
		return true
	}
	return rb.Enqueue(v)
}

// Pop reads the next item in the buffer.
// We retry in case of some conflict with other goroutine.
func (rb *RingBuffer[Value]) Dequeue() (Value, bool) {
	originalState := rb.state.Load()
	state := newState(originalState)
	var v Value
	if state.Empty() {
		return v, false
	}
	el := rb.elements[state.head%rb.capacity]
	state.head++
	state.full = false
	val := el.Load()
	if rb.state.CompareAndSwap(originalState, state.Uint64()) {
		if val != nil {
			v = *val
		}
		return v, true
	}
	// in case we have some conflict with another goroutine, retry.
	return rb.Dequeue()
}
