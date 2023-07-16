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

func TestReactor_NewWithArgs(t *testing.T) {
	r := New(
		WithCallbacksDemux(NewDemux[event[mockEvent]]()),
		WithEventsDemux(NewDemux[event[mockEvent]]()),
		WithTimes[mockEvent](time.Second, time.Minute),
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5)
	defer cancel()

	go func() {
		_ = r.Start(ctx)
	}()
	defer func() {
		_ = r.Close()
	}()
}

func TestReactor_CallbackTimeout(t *testing.T) {
	pctx, pcancel := context.WithCancel(context.Background())
	defer pcancel()

	timeout := time.Second
	r := New(WithTimes[mockEvent](time.Second/4, timeout))
	go func() {
		_ = r.Start(pctx)
	}()
	defer func() {
		_ = r.Close()
	}()
	r.AddHandler("test-1", func(me mockEvent) bool {
		return len(me.name) > 0
	}, 1, func(me mockEvent, callback func(mockEvent, error)) {
		me.count++
		go func(me mockEvent) {
			<-time.After(timeout * 2)
			callback(me, nil)
		}(me)
	})
	defer r.RemoveHandler("test-1")

	res, err := r.EnqueueWait(pctx, mockEvent{
		name: "hello timeout",
	})
	require.Error(t, err)
	require.Equal(t, res, mockEvent{})
}

func TestReactor_Sanity(t *testing.T) {
	pctx, pcancel := context.WithCancel(context.Background())
	defer pcancel()

	r := New(WithTimes[mockEvent](time.Second/4, time.Second))
	go func() {
		_ = r.Start(pctx)
	}()
	defer func() {
		_ = r.Close()
	}()

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
		me.count++
		go func(me mockEvent) {
			<-time.After(time.Millisecond * 5)
			callback(me, errors.New("test-error"))
		}(me)
	})
	defer r.RemoveHandler("test-err")

	callbackCounter := atomic.Int32{}
	r.AddCallback("test-all", func(me mockEvent) bool {
		return len(me.name) > 0
	}, 1, func(me mockEvent) {
		if len(me.name) > 0 {
			require.Greater(t, me.count, int32(0))
		}
		t.Logf("got event %+v", me)
		callbackCounter.Add(1)
	})
	defer r.RemoveCallback("test-all")

	n := 4
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	go func(n int) {
		n = n - 2 // because we are doing additional enqueues after this loop
		for n > 0 {
			r.Enqueue(mockEvent{
				name: fmt.Sprintf("test-event-%d", n),
			})
			n--
		}
		r.Enqueue(mockEvent{
			name: "errored",
		})
		res, err := r.EnqueueWait(ctx, mockEvent{
			name: fmt.Sprintf("test-event-%d", n),
		})
		require.NoError(t, err)
		require.Greater(t, res.count, int32(0))
	}(n)

	for callbackCounter.Load() < int32(n) && ctx.Err() == nil {
		time.Sleep(time.Millisecond * 10)
	}
	<-time.After(time.Second) // to ensure there are no leftovers
	require.Equal(t, callbackCounter.Load(), int32(n))
	require.NoError(t, r.Close())
}

func TestReactor_EventID(t *testing.T) {
	r := New[[]byte]().(*reactor[[]byte])

	id := r.genID([]byte("hello-test"))
	require.Len(t, id, 18)
	require.Equal(t, id, IDFromString(id.String()), "id encoding failed")
	require.Equal(t, ID{}, IDFromString(""), "empty id encoding failed")
	require.Equal(t, ID(nil), IDFromString("`^"), "invalid id encoding failed")
}
