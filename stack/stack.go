package stack

import (
	"errors"
	"sync/atomic"
)

var (
	ErrStackOverflow = errors.New("stack overflow")
)

// element is an item in the stack.
// It contains a pointer to the value + the next element.
type element[Value any] struct {
	value atomic.Pointer[Value]
	next  atomic.Pointer[element[Value]]
}

// Stack is a lock-free stack implementation,
// based on atomic compare-and-swap operations.
type Stack[Value any] struct {
	head  atomic.Pointer[element[Value]]
	size  atomic.Int32
	limit int32
}

// New creates a new lock-free stack.
func New[Value any](limit int) *Stack[Value] {
	return &Stack[Value]{
		head:  atomic.Pointer[element[Value]]{},
		size:  atomic.Int32{},
		limit: int32(limit),
	}
}

// Push adds a new value to the stack.
// It keeps retrying in case of conflict with concurrent Pop()/Push() operations.
func (s *Stack[Value]) Push(value Value) error {
	if s.IsFull() {
		return ErrStackOverflow
	}

	e := &element[Value]{}
	e.value.Store(&value)

	for {
		h := s.head.Load()
		e.next.Store(h)
		if s.head.CompareAndSwap(h, e) {
			_ = s.size.Add(1)
			break
		}
	}
	return nil
}

// Pop removes the next value from the stack.
// returns false in case the stack wasn't changed,
// which could happen if the stack is empty or
// if there was a conflict with concurrent Pop()/Push() operation.
func (s *Stack[Value]) Pop() (*Value, bool) {
	h := s.head.Load()
	if h == nil {
		return nil, false
	}
	next, value := (*h).next.Load(), (*h).value.Load()
	changed := s.head.CompareAndSwap(h, next)
	if changed {
		_ = s.size.Add(-1)
	}
	return value, changed
}

// Range iterates over the stack, accepts a custom iterator that returns true to stop.
func (s *Stack[Value]) Range(iterator func(val Value) bool) {
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

// Len returns the number of items in the stack.
func (s *Stack[Value]) Len() int {
	return int(s.size.Load())
}

// Len returns the number of items in the stack.
func (s *Stack[Value]) IsFull() bool {
	return s.size.Load() == s.limit
}
