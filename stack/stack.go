package stack

import "sync/atomic"

// element is an item in the stack.
// It contains a pointer to the value + the next element.
type element[Value any] struct {
	value atomic.Pointer[Value]
	next  atomic.Pointer[*element[Value]]
}

// Stack is a lock-free stack implementation,
// based on atomic compare-and-swap operations.
type Stack[Value any] struct {
	head atomic.Pointer[*element[Value]]
}

// New creates a new lock-free stack.
func New[Value any]() *Stack[Value] {
	return &Stack[Value]{}
}

// Push adds a new value to the stack.
// It keeps retrying in case of conflict with concurrent Pop()/Push() operations.
func (s *Stack[Value]) Push(value Value) {
	e := &element[Value]{}
	e.value.Store(&value)

	for {
		h := s.head.Load()
		e.next.Store(h)
		if s.head.CompareAndSwap(h, &e) {
			break
		}
	}
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
	changed := s.head.CompareAndSwap(h, (*h).next.Load())
	return (*h).value.Load(), changed
}

// Range iterates over the stack, accepts a custom iterator that returns true to stop.
func (s *Stack[Value]) Range(iterator func(val Value) bool) {
	currentP := s.head.Load()
	if currentP == nil {
		return
	}
	current := *currentP
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
		currentP := current.next.Load()
		if currentP == nil {
			return
		}
		current = *currentP
	}
}

// Len returns the number of items in the stack.
// Utilizes Range() to count the items.
func (s *Stack[Value]) Len() int {
	var counter int
	s.Range(func(val Value) bool {
		counter++
		return false
	})
	return counter
}
