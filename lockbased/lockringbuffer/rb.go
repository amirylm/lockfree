package lockringbuffer

import (
	"sync"

	"github.com/amirylm/lockfree/common"
)

func New[Value any](c int) common.DataStructure[Value] {
	rb := &LockRingBuffer[Value]{
		lock:     &sync.RWMutex{},
		data:     make([]Value, c),
		capacity: uint32(c),
	}

	return rb
}

// LockRingBuffer is a ring buffer that uses rw mutex to provide thread safety
type LockRingBuffer[Value any] struct {
	lock *sync.RWMutex

	data []Value

	state    ringBufferState
	capacity uint32
}

func (rb *LockRingBuffer[V]) Empty() bool {
	rb.lock.RLock()
	defer rb.lock.RUnlock()

	return rb.state.Empty()
}

func (rb *LockRingBuffer[V]) Full() bool {
	rb.lock.RLock()
	defer rb.lock.RUnlock()

	return rb.state.Full()
}

// Push adds a new item to the buffer.
func (rb *LockRingBuffer[V]) Push(v V) bool {
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
func (rb *LockRingBuffer[V]) Pop() (V, bool) {
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

func (rb *LockRingBuffer[Value]) Size() int {
	rb.lock.RLock()
	defer rb.lock.RUnlock()

	state := rb.state
	return int(state.tail - state.head)
}
