package gochan

import (
	"github.com/amirylm/go-options"
	"github.com/amirylm/lockfree/core"
)

type GoChanQ[Value any] struct {
	cn       chan Value
	capacity int
}

func New[Value any](opts ...options.Option[core.Options]) core.Queue[Value] {
	o := options.Apply(nil, opts...)
	gc := &GoChanQ[Value]{
		capacity: int(o.Capacity()),
	}
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
