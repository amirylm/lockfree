package core

import (
	"errors"
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

// Stack is the interface for working with a stack.
type Stack[T any] interface {
	Push(T) bool
	Pop() (T, bool)

	DataStructureBase
}
