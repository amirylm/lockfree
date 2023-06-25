package common

import (
	"context"
	"errors"
	"runtime"
)

var (
	ErrOverflow = errors.New("data overflow")
)

// DataStructureFactory is a factory function for creating instances of ds
type DataStructureFactory[Value any] func(capacity int) DataStructure[Value]

// DataStructure is implemented by all the data structures in this project.
type DataStructure[Value any] interface {
	Push(Value) bool
	Pop() (Value, bool)
	Size() int
	Empty() bool
	Full() bool
}

// Push is used as a safe function to write values into the ds,
// ensuring that even if the q is full, we'll try to write the item
// until the context is done.
func Push[Value any](ctx context.Context, ds DataStructure[Value], val Value) bool {
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
func Pop[Value any](ctx context.Context, ds DataStructure[Value]) (Value, bool) {
	element, ok := ds.Pop()
	for !ok && ctx.Err() == nil {
		runtime.Gosched()
		element, ok = ds.Pop()
	}
	return element, true
}
