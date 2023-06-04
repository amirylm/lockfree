package common

import (
	"context"
	"runtime"
)

type Handler[Value any] func(Value) bool

func Channel[Value any](ctx context.Context, ds DataStructure[Value], reader Handler[Value], errHandler Handler[error]) Handler[Value] {
	go func() {
		for ctx.Err() == nil {
			if ds.Empty() {
				runtime.Gosched()
				continue
			}
			val, ok := ds.Pop()
			if !ok {
				continue
			}
			if !reader(val) {
				return
			}
		}
	}()

	return ds.Push
}
