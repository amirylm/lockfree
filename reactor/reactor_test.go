package reactor

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockEvent struct {
	count int32
	name  string
}

func TestReactor_Sanity(t *testing.T) {
	r := New[mockEvent]()

	r.AddHandler("test-1", func(me mockEvent) bool {
		return len(me.name) > 0 && me.name != "errored"
	}, 1, func(me mockEvent, callback func(mockEvent, error)) {
		me.count++
		go func(me mockEvent) {
			<-time.After(time.Millisecond * 5)
			callback(me, nil)
		}(me)
	})
	defer r.RemoveHandler("test-1")

	r.AddHandler("test-err", func(me mockEvent) bool {
		return me.name == "errored"
	}, 1, func(me mockEvent, callback func(mockEvent, error)) {
		go func(me mockEvent) {
			me.count++
			<-time.After(time.Millisecond * 20)
			callback(me, errors.New("test-error"))
		}(me)
	})
	defer r.RemoveHandler("test-err")

	callbackCounter := atomic.Int32{}
	r.AddCallback("test-all", func(me mockEvent) bool {
		return true
	}, 1, func(me mockEvent) {
		callbackCounter.Add(1)
	})
	defer r.RemoveCallback("test-all")

	n := 4
	go func(n int) {
		for n > 0 {
			r.Enqueue(mockEvent{
				name: fmt.Sprintf("test-event-%d", n),
			})
			n--
		}
	}(n * 2)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	for callbackCounter.Load() < int32(n) && ctx.Err() == nil {
		time.Sleep(time.Millisecond * 10)
	}
	require.NoError(t, ctx.Err())
}
