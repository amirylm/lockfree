package common

import (
	"context"
	"runtime"
)

type LockFreeCollection[Value any] interface {
	Push(Value) error
	Pop() (Value, bool)
	Size() int
	Empty() bool
	Full() bool
}

type Handler[Value any] func(Value) error

func Channel[Value any](ctx context.Context, collection LockFreeCollection[Value], reader Handler[Value], errHandler func(error) bool) Handler[Value] {
	checkErr := func(err error) bool {
		if err != nil {
			if errHandler == nil {
				return true
			}
			return errHandler(err)
		}
		return false
	}

	go func() {
		for ctx.Err() == nil {
			if collection.Empty() {
				runtime.Gosched()
				continue
			}
			val, ok := collection.Pop()
			if !ok {
				continue
			}
			err := reader(val)
			if checkErr(err) {
				return
			}
		}
	}()

	return func(v Value) error {
		return collection.Push(v)
	}
}
