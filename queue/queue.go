package queue

import (
	"sync/atomic"

	"github.com/amirylm/go-options"
	"github.com/amirylm/lockfree/core"
)

type element[Value any] struct {
	value Value
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

func WithCapacity[Value any](c int) options.Option[Queue[Value]] {
	return func(q *Queue[Value]) {
		q.capacity = int32(c)
	}
}

func New[Value any](opts ...options.Option[Queue[Value]]) core.Queue[Value] {
	var e = element[Value]{}
	q := &Queue[Value]{
		size: atomic.Int32{},
	}
	_ = options.Apply(q, opts...)
	q.head.Store(&e)
	q.tail.Store(&e)
	return q
}

func (q *Queue[Value]) Enqueue(v Value) bool {
	if q.Full() {
		return false
	}
	e := &element[Value]{value: v}
	for {
		t := q.tail.Load()
		tn := t.next.Load()
		if tn == nil { // tail next is nil: assign element and shift pointer
			if t.next.CompareAndSwap(tn, e) {
				q.tail.CompareAndSwap(t, e)
				q.size.Add(1)
				return true
			}
		} else {
			q.tail.CompareAndSwap(t, tn) // reassign tail pointer to next
		}
	}
}

func (q *Queue[Value]) Dequeue() (Value, bool) {
	var v Value
	for {
		t := q.tail.Load()
		h := q.head.Load()
		next := h.next.Load()
		if h != t { // element exists
			v := next.value
			if q.head.CompareAndSwap(h, next) { // set head to next
				q.size.Add(-1)
				return v, true
			}
			// head and tail are equal (either empty or single entity)
		} else {
			if next == nil {
				return v, false
			}
			q.tail.CompareAndSwap(t, next)
		}
	}
}

func (q *Queue[Value]) Size() int {
	return int(q.size.Load())
}

func (q *Queue[Value]) Empty() bool {
	return q.head.Load() == q.tail.Load()
}

func (q *Queue[Value]) Full() bool {
	return q.size.Load() == q.capacity
}

func (q *Queue[Value]) Range(iterator func(val Value) bool) {
	// head pointer never holds a value, only denotes start
	h := q.head.Load()
	if h == nil {
		return
	}
	current := h.next.Load()
	for current != nil {
		v := current.value
		if iterator(v) {
			return
		}
		current = current.next.Load()
	}
}
