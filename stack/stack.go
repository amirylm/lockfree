package stack

import (
	"sync/atomic"

	"github.com/amirylm/lockfree/core"
)

// element is an item in the stack.
// It contains a pointer to the value + the next element.
type element[Value any] struct {
	value atomic.Pointer[Value]
	next  atomic.Pointer[element[Value]]
}

// LLStack is a lock-free stack implemented with linked list,
// based on atomic compare-and-swap operations.
type LLStack[Value any] struct {
	head     atomic.Pointer[element[Value]]
	size     atomic.Int32
	capacity int32
}

// New creates a new lock-free stack.
func New[Value any](capacity int) core.Stack[Value] {
	return &LLStack[Value]{
		head:     atomic.Pointer[element[Value]]{},
		size:     atomic.Int32{},
		capacity: int32(capacity),
	}
}

// Push adds a new value to the stack.
// It keeps retrying in case of conflict with concurrent Pop()/Push() operations.
func (s *LLStack[Value]) Push(value Value) bool {
	e := &element[Value]{}
	e.value.Store(&value)

	for !s.Full() {
		h := s.head.Load()
		e.next.Store(h)
		if s.head.CompareAndSwap(h, e) {
			_ = s.size.Add(1)
			return true
		}
	}
	return false
}

// Pop removes the next value from the stack.
func (s *LLStack[Value]) Pop() (Value, bool) {
	var val Value
	h := s.head.Load()
	if h == nil {
		return val, false
	}
	next, value := (*h).next.Load(), (*h).value.Load()

	if changed := s.head.CompareAndSwap(h, next); changed {
		_ = s.size.Add(-1)
		if value != nil {
			val = *value
		}
		return val, true
	}

	return val, false
}

// Range iterates over the stack, accepts a custom iterator that returns true to stop.
func (s *LLStack[Value]) Range(iterator func(val Value) bool) {
	current := s.head.Load()
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

// Size returns the number of items in the stack.
func (s *LLStack[Value]) Size() int {
	return int(s.size.Load())
}

// Len returns the number of items in the stack.
func (s *LLStack[Value]) Full() bool {
	return s.size.Load() == s.capacity
}

// Len returns the number of items in the stack.
func (s *LLStack[Value]) Empty() bool {
	return s.size.Load() == 0
}
