package core

import (
	"context"
	"errors"
	"runtime"
)

var (
	ErrOverflow = errors.New("data overflow")
)

// DataStructure is the base interface for all data structures.
type DataStructureBase interface {
	Size() int
	Empty() bool
	Full() bool
}

// Queue is the interface for working with a queue.
type Queue[T any] interface {
	Enqueue(T) bool
	Dequeue() (T, bool)

	DataStructureBase
}

// Enqueue is used as a safe function to write values into queues,
// ensuring that even if the q is full, we'll try to write the item
// until the context is done.
func Enqueue[T any](ctx context.Context, ds Queue[T], val T) bool {
	for !ds.Enqueue(val) {
		if ctx.Err() != nil {
			return false
		}
		runtime.Gosched()
	}
	return true
}

// Pop is a safe function to read values from the ds,
// until the context is done.
func Dequeue[T any](ctx context.Context, ds Queue[T]) (T, bool) {
	element, ok := ds.Dequeue()
	for !ok && ctx.Err() == nil {
		runtime.Gosched()
		element, ok = ds.Dequeue()
	}
	return element, true
}

// Stack is the interface for working with a stack.
type Stack[T any] interface {
	Push(T) bool
	Pop() (T, bool)

	DataStructureBase
}

// Push is used as a safe function to write values into the ds,
// ensuring that even if the q is full, we'll try to write the item
// until the context is done.
func Push[T any](ctx context.Context, ds Stack[T], val T) bool {
	for !ds.Push(val) {
		if ctx.Err() != nil {
			return false
		}
		runtime.Gosched()
	}
	return true
}

// Pop is a safe function to read values from the ds,
// until the context is done.
func Pop[T any](ctx context.Context, ds Stack[T]) (T, bool) {
	element, ok := ds.Pop()
	for !ok && ctx.Err() == nil {
		runtime.Gosched()
		element, ok = ds.Pop()
	}
	return element, true
}
