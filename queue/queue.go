package queue

import (
	"sync/atomic"

	"github.com/amirylm/lockfree/common"
)

type element[T any] struct {
	value T
	next  atomic.Pointer[element[T]]
}

type Queue[T any] struct {
	head     atomic.Pointer[element[T]]
	tail     atomic.Pointer[element[T]]
	size     atomic.Int32
	capacity int32
}

func newElement[T any](v T) *element[T] {
	return &element[T]{value: v}
}

func New[T any](capacity int) common.DataStructure[T] {
	var head atomic.Pointer[element[T]]
	var tail atomic.Pointer[element[T]]
	var e = element[T]{}
	head.Store(&e)
	tail.Store(&e)
	return &Queue[T]{
		head: head,
		tail: tail,
	}
}

func (q *Queue[T]) Size() int {
	return int(q.size.Load())
}

func (q *Queue[T]) Push(v T) bool {
	e := newElement(v)
	for {
		tail := q.tail.Load()
		next := tail.next.Load()
		if tail == q.tail.Load() {
			if next == nil {
				if tail.next.CompareAndSwap(next, e) {
					q.tail.CompareAndSwap(tail, e)
					q.size.Add(1)
					return true
				}
			} else {
				q.tail.CompareAndSwap(tail, next)
			}
		}
	}
}

func (q *Queue[T]) Pop() (T, bool) {
	var t T
	for {
		head := q.head.Load()
		next := head.next.Load()
		tail := q.tail.Load()
		if head == q.head.Load() {
			if head != tail {
				v := next.value
				if q.head.CompareAndSwap(head, next) {
					return v, true
				}
				// head and tail are equal (either empty or single entity)
			} else {
				if next == nil {
					return t, false
				}
				q.tail.CompareAndSwap(tail, next)
			}
		}
	}
}

func (q *Queue[T]) Empty() bool {
	return q.head.Load() == q.tail.Load()
}

func (q *Queue[T]) Full() bool {
	return q.size.Load() == q.capacity
}
