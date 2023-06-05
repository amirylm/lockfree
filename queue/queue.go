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

// Queue is a lock-free queue implementation,
// based on atomic compare-and-swap operations.
type Queue[Value any] struct {
	head atomic.Pointer[element[Value]]
	tail atomic.Pointer[element[Value]]

	capacity int
}

// New creates a new lock-free queue.
func New[Value any](capacity int) common.DataStructure[Value] {
	return &Queue[Value]{
		head:     atomic.Pointer[element[Value]]{},
		tail:     atomic.Pointer[element[Value]]{},
		capacity: capacity,
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
			break
		}
		(*t).next.Store(e)
		if q.tail.CompareAndSwap(t, e) {
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
	h := q.head.Load()
	if h == nil {
		return val, false
	}
	next, value := (*h).next.Load(), (*h).value.Load()
	changed := q.head.CompareAndSwap(h, next)
	if value != nil {
		val = *value
	}
	return val, changed
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
		stop := iterator(val)
		if stop {
			return
		}
		current = current.next.Load()
	}
}

// Size returns the number of items in the queue.
// Utilizes Range() to count the items.
func (q *Queue[Value]) Size() int {
	var counter int
	q.Range(func(val Value) bool {
		counter++
		return false
	})
	return counter
}

func (q *Queue[Value]) Full() bool {
	return q.Size() == int(q.capacity)
}

func (q *Queue[Value]) Empty() bool {
	return q.Size() == 0
}
