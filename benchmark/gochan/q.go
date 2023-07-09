package gochan

import (
	"github.com/amirylm/go-options"
	"github.com/amirylm/lockfree/core"
)

type GoChanQ[Value any] struct {
	cn       chan Value
	capacity int
}

func WithCapacity[Value any](c int) options.Option[GoChanQ[Value]] {
	return func(q *GoChanQ[Value]) {
		q.capacity = c
	}
}

func New[Value any](opts ...options.Option[GoChanQ[Value]]) core.Queue[Value] {
	gc := &GoChanQ[Value]{}

	_ = options.Apply(gc, opts...)
	gc.cn = make(chan Value, gc.capacity)

	return gc
}

func (q *GoChanQ[Value]) Enqueue(v Value) bool {
	select {
	case q.cn <- v:
		return true
	default:
	}
	return false
}

func (q *GoChanQ[Value]) Dequeue() (Value, bool) {
	v, ok := <-q.cn
	return v, ok
}

func (q *GoChanQ[Value]) Size() int {
	return len(q.cn)
}

func (q *GoChanQ[Value]) Empty() bool {
	return q.Size() == 0
}

func (q *GoChanQ[Value]) Full() bool {
	return q.Size() == q.capacity
}
