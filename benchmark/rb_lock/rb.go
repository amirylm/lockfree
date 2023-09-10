package rb_lock

import (
	"sync"

	"github.com/amirylm/go-options"
	"github.com/amirylm/lockfree/core"
)

func New[Value any](opts ...options.Option[core.Options]) core.Queue[Value] {
	o := options.Apply(nil, opts...)
	rb := &RingBufferLock[Value]{
		lock:     &sync.RWMutex{},
		capacity: uint32(o.Capacity()),
	}
	rb.data = make([]Value, rb.capacity)

	return rb
}

type ringBufferState struct {
	head, tail uint32
	full       bool
}

func (state ringBufferState) Empty() bool {
	return !state.full && state.head == state.tail
}

func (state ringBufferState) Full() bool {
	return state.full
}

// RingBufferLock is a ring buffer that uses rw mutex to provide thread safety
type RingBufferLock[Value any] struct {
	lock *sync.RWMutex

	data []Value

	state    ringBufferState
	capacity uint32
}

func (rb *RingBufferLock[V]) Empty() bool {
	rb.lock.RLock()
	defer rb.lock.RUnlock()

	return rb.state.Empty()
}

func (rb *RingBufferLock[V]) Full() bool {
	rb.lock.RLock()
	defer rb.lock.RUnlock()

	return rb.state.Full()
}

// Push adds a new item to the buffer.
func (rb *RingBufferLock[V]) Enqueue(v V) bool {
	rb.lock.Lock()
	defer rb.lock.Unlock()

	state := rb.state
	if state.full {
		return false
	}
	i := state.tail % rb.capacity
	state.tail++
	state.full = (state.tail%rb.capacity == state.head%rb.capacity)
	rb.data[i] = v
	rb.state = state

	return true
}

// Enqueue pops the next item in the buffer.
func (rb *RingBufferLock[V]) Dequeue() (V, bool) {
	rb.lock.Lock()
	defer rb.lock.Unlock()

	var empty V
	state := rb.state
	if !state.full && (state.tail%rb.capacity == state.head%rb.capacity) {
		return empty, false
	}
	i := state.head % rb.capacity
	v := rb.data[i]
	state.head++
	state.full = false
	rb.data[i] = empty
	rb.state = state

	return v, true
}

func (rb *RingBufferLock[Value]) Size() int {
	rb.lock.RLock()
	defer rb.lock.RUnlock()

	state := rb.state
	return int(state.tail - state.head)
}
