package stack

import (
	"github.com/amirylm/lockfree/core"
)

type QueueAdapter[T any] struct {
	s core.Stack[T]
}

func NewQueueAdapter[T any](capacity int) core.Queue[T] {
	return &QueueAdapter[T]{
		s: New[T](core.WithCapacity(capacity)),
	}
}

func (q *QueueAdapter[T]) Enqueue(v T) bool {
	return q.s.Push(v)
}

func (q *QueueAdapter[T]) Dequeue() (T, bool) {
	return q.s.Pop()
}

func (q *QueueAdapter[T]) Size() int {
	return q.s.Size()
}

func (q *QueueAdapter[T]) Empty() bool {
	return q.s.Empty()
}

func (q *QueueAdapter[T]) Full() bool {
	return q.s.Full()
}
