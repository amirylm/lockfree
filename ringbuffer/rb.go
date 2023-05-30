package ringbuffer

import (
	"errors"
	"runtime"
	"sync/atomic"
)

// NOTE: WIP

var (
	ErrBufferIsEmpty  = errors.New("buffer is empty")
	ErrBufferOverflow = errors.New("buffer overflow")
)

func New[V any](c int) *RingBuffer[V] {
	rb := &RingBuffer[V]{
		data:     make([]V, c),
		capacity: uint32(c),
	}

	return rb
}

// RingBuffer is a lock-free ring buffer implementation.
// NOTE: WIP
type RingBuffer[V any] struct {
	data []V

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
		return ErrBufferOverflow
	}
	i := state.tail % rb.capacity
	state.tail++
	state.full = (state.tail%rb.capacity == state.head%rb.capacity)
	orig := rb.data[i]
	rb.data[i] = v
	if !rb.state.CompareAndSwap(originalState, state.Uint64()) {
		rb.data[i] = orig
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
	var empty V
	if state.IsEmpty() {
		return empty, false
	}
	i := state.head % rb.capacity
	v := rb.data[i]
	state.head++
	state.full = false
	rb.data[i] = empty
	if !rb.state.CompareAndSwap(originalState, state.Uint64()) {
		// in case we have some conflict with another goroutine, revert and retry.
		rb.data[i] = v
		runtime.Gosched()
		return rb.Pop()
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
