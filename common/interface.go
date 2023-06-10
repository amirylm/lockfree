package common

import (
	"context"
	"errors"
	"runtime"
)

var (
	ErrOverflow = errors.New("data overflow")
)

type DataStructureFactory[Value any] func(capacity int) DataStructure[Value]

type DataStructure[Value any] interface {
	Push(Value) bool
	Pop() (Value, bool)
	Size() int
	Empty() bool
	Full() bool
}

func Push[Value any](ctx context.Context, ds DataStructure[Value], val Value) bool {
	for !ds.Push(val) {
		if ctx.Err() != nil {
			return false
		}
		runtime.Gosched()
	}
	return true
}

func Pop[Value any](ctx context.Context, ds DataStructure[Value]) (Value, bool) {
	element, ok := ds.Pop()
	for !ok && ctx.Err() == nil {
		runtime.Gosched()
		element, ok = ds.Pop()
	}
	return element, true
}
