package gochan

import "github.com/amirylm/lockfree/common"

type GoChanQ[Value any] struct {
	cn       chan Value
	capacity int
}

func New[Value any](capacity int) common.DataStructure[Value] {
	return &GoChanQ[Value]{
		cn:       make(chan Value, capacity),
		capacity: capacity,
	}
}

func (q *GoChanQ[Value]) Push(v Value) bool {
	select {
	case q.cn <- v:
		return true
	default:
	}
	return false
}

func (q *GoChanQ[Value]) Pop() (Value, bool) {
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
