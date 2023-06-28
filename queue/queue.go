package queue

import (
	"sync/atomic"

	"github.com/amirylm/lockfree/common"
)

type element[Value any] struct {
	value Value
	next  atomic.Pointer[element[Value]]
}

type Queue[Value any] struct {
	head     atomic.Pointer[element[Value]]
	tail     atomic.Pointer[element[Value]]
	size     atomic.Int32
	capacity int32
}

func New[Value any](capacity int) common.DataStructure[Value] {
	var e = element[Value]{}
	var head atomic.Pointer[element[Value]]
	var tail atomic.Pointer[element[Value]]
	head.Store(&e)
	tail.Store(&e)
	return &Queue[Value]{
		head: head,
		tail: tail,
	}
}

func initElement[Value any](v Value) *element[Value] {
	return &element[Value]{value: v}
}

func (q *Queue[Value]) Size() int {
	return int(q.size.Load())
}

func (q *Queue[Value]) Push(v Value) bool {
	e := initElement(v)
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

func (q *Queue[Value]) Pop() (Value, bool) {
	var e Value
	for {
		t := q.tail.Load()
		h := q.head.Load()
		hn := h.next.Load()
		if h != t { // element exists
			v := hn.value
			if q.head.CompareAndSwap(h, hn) { // set head to next
				return v, true
			}
			// head and tail are equal (either empty or single entity)
		} else {
			if hn == nil {
				return e, false
			}
			q.tail.CompareAndSwap(t, hn)
		}
	}
}

func (q *Queue[Value]) Empty() bool {
	return q.head.Load() == q.tail.Load()
}

func (q *Queue[Value]) Full() bool {
	return q.size.Load() == q.capacity
}
