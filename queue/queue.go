package queue

import (
	"sync/atomic"
)

// element is an item in the ueue.
// It contains a pointer to the value + the next element.
type element[Value any] struct {
	value atomic.Pointer[Value]
	next  atomic.Pointer[*element[Value]]
}

// Queue is a lock-free queue implementation,
// based on atomic compare-and-swap operations.
type Queue[Value any] struct {
	head atomic.Pointer[*element[Value]]
	tail atomic.Pointer[*element[Value]]
}

// New creates a new lock-free queue.
func New[Value any]() *Queue[Value] {
	return &Queue[Value]{}
}

// Push adds a new value to the end of the queue.
// It keeps retrying in case of conflict with concurrent Pop()/Push() operations.
func (q *Queue[Value]) Push(value Value) {
	e := &element[Value]{}
	e.value.Store(&value)

	for {
		t := q.tail.Load()
		if t == nil {
			q.head.Store(&e)
			q.tail = q.head
			break
		}
		(*t).next.Store(&e)
		if q.tail.CompareAndSwap(t, &e) {
			break
		}
	}
}

// Pop removes the first value from the queue.
// returns false in case the queue wasn't changed,
// which could happen if the queue is empty or
// if there was a conflict with concurrent Pop()/Push() operation.
func (q *Queue[Value]) Pop() (*Value, bool) {
	h := q.head.Load()
	if h == nil {
		return nil, false
	}
	changed := q.head.CompareAndSwap(h, (*h).next.Load())
	return (*h).value.Load(), changed
}

// Range iterates over the queue, accepts a custom iterator that returns true to stop.
func (q *Queue[Value]) Range(iterator func(val Value) bool) {
	currentP := q.head.Load()
	if currentP == nil {
		return
	}
	current := *currentP
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
		currentP := current.next.Load()
		if currentP == nil {
			return
		}
		current = *currentP
	}
}

// Len returns the number of items in the queue.
// Utilizes Range() to count the items.
func (q *Queue[Value]) Len() int {
	var counter int
	q.Range(func(val Value) bool {
		counter++
		return false
	})
	return counter
}
