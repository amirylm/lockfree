package stack

import (
	"github.com/amirylm/go-options"
	"github.com/amirylm/lockfree/core"
)

type QueueAdapter[T any] struct {
	s core.Stack[T]
}

func WithCapacity[Value any](c int) options.Option[LLStack[Value]] {
	return func(s *LLStack[Value]) {
		s.capacity = int32(c)
	}
}

func NewQueueAdapter[T any](capacity int) core.Queue[T] {
	return &QueueAdapter[T]{
		s: New[T](WithCapacity[T](capacity)),
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
