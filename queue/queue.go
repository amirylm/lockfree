package queue

import (
	"sync/atomic"

	"github.com/amirylm/lockfree/common"
)

// element is an item in the ueue.
// It contains a pointer to the value + the next element.
type element[Value any] struct {
	value atomic.Pointer[Value]
	next  atomic.Pointer[element[Value]]
}

// Queue is a lock-free queue implemented with linked list,
// based on atomic compare-and-swap operations.
type Queue[Value any] struct {
	head     atomic.Pointer[element[Value]]
	tail     atomic.Pointer[element[Value]]
	size     atomic.Int32
	capacity int32
}

// New creates a new lock-free queue.
func New[Value any](capacity int) common.DataStructure[Value] {
	return &Queue[Value]{
		head:     atomic.Pointer[element[Value]]{},
		tail:     atomic.Pointer[element[Value]]{},
		size:     atomic.Int32{},
		capacity: int32(capacity),
	}
}

// Push adds a new value to the end of the queue.
// It keeps retrying in case of conflict with concurrent Pop()/Push() operations.
func (q *Queue[Value]) Push(value Value) bool {
	if q.Full() {
		return false
	}
	e := &element[Value]{}
	e.value.Store(&value)

	for {
		t := q.tail.Load()
		if t == nil {
			q.head.Store(e)
			q.tail.Store(q.head.Load())
			q.size.Add(1)
			break
		}
		(*t).next.Store(e)
		if q.tail.CompareAndSwap(t, e) {
			q.size.Add(1)
			break
		}
	}
	return true
}

// Pop removes the first value from the queue.
// returns false in case the queue wasn't changed,
// which could happen if the queue is empty or
// if there was a conflict with concurrent Pop()/Push() operation.
func (q *Queue[Value]) Pop() (Value, bool) {
	var val Value
	if q.Empty() {
		return val, false
	}
	h := q.head.Load()
	if h == nil {
		return val, false
	}
	next, value := (*h).next.Load(), (*h).value.Load()
	if q.head.CompareAndSwap(h, next) {
		q.size.Add(-1)
		if value != nil {
			val = *value
		}
		return val, true
	}
	return q.Pop()
}

// Range iterates over the queue, accepts a custom iterator that returns true to stop.
func (q *Queue[Value]) Range(iterator func(val Value) bool) {
	current := q.head.Load()
	for current != nil {
		var val Value
		valp := current.value.Load()
		if valp != nil {
			val = *valp
		}
		if iterator(val) {
			return
		}
		current = current.next.Load()
	}
}

// Size returns the number of items in the stack.
func (q *Queue[Value]) Size() int {
	return int(q.size.Load())
}

// Len returns the number of items in the stack.
func (q *Queue[Value]) Full() bool {
	return q.size.Load() == q.capacity
}

// Len returns the number of items in the stack.
func (q *Queue[Value]) Empty() bool {
	return q.size.Load() == 0
}
