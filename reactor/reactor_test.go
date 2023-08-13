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

type mockEventData struct {
	Count int32
	name  string
}

type ReactiveServiceImpl struct {
	SelectLogic func(event Event[mockEventData]) bool
	HandleLogic func(event Event[mockEventData], callback func(mockEventData, error))
}

func (rs *ReactiveServiceImpl) Select(event Event[mockEventData]) bool {
	// Call the logic function for Select with the event as an argument
	return rs.SelectLogic(event)
}

func (rs *ReactiveServiceImpl) Handle(event Event[mockEventData], callback func(mockEventData, error)) {
	// Call the logic function for Handle with the event and callback as arguments
	rs.HandleLogic(event, callback)
}

type CallbackServiceImpl struct {
	SelectLogic func(event Event[mockEventData]) bool
	HandleLogic func(callback Event[mockEventData])
}

func (cs *CallbackServiceImpl) Select(event Event[mockEventData]) bool {
	// Call the logic function for Select with the event as an argument
	return cs.SelectLogic(event)
}

func (cs *CallbackServiceImpl) Handle(callback Event[mockEventData]) {
	// Call the logic function for Handle with the event and callback as arguments
	cs.HandleLogic(callback)
}

func TestReactor_NewWithArgs(t *testing.T) {
	r := New(
		WithCallbacksDemux[mockEventData](NewDemux[Event[mockEventData]]()),
		WithEventsDemux[mockEventData, mockEventData](NewDemux[Event[mockEventData]]()),
		WithTimes[mockEventData, mockEventData](time.Second, time.Minute),
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
	r := New(WithTimes[mockEventData, mockEventData](time.Second/4, timeout))
	rs := &ReactiveServiceImpl{
		SelectLogic: func(e Event[mockEventData]) bool {
			return true
		},
		HandleLogic: func(e Event[mockEventData], callback func(mockEventData, error)) {
			// no callback -> timeout
			// callback(e.Data, nil)
		},
	}
	go func() {
		_ = r.Start(pctx)
	}()
	defer func() {
		_ = r.Close()
	}()
	r.AddHandler("test-1", rs, 1)
	defer r.RemoveHandler("test-1")

	res, err := r.EnqueueWait(pctx, mockEventData{
		name: "hello timeout",
	})
	require.Error(t, err)
	require.Equal(t, res, mockEventData{})
}

func TestReactor_Sanity(t *testing.T) {
	pctx, pcancel := context.WithCancel(context.Background())
	defer pcancel()

	r := New(WithTimes[mockEventData, mockEventData](time.Second/4, time.Second*2))
	rs := &ReactiveServiceImpl{
		SelectLogic: func(e Event[mockEventData]) bool {
			return len(e.Data.name) > 0 && e.Data.name != "errored"
		},
		HandleLogic: func(e Event[mockEventData], callback func(mockEventData, error)) {
			e.Data.Count++
			go func(me mockEventData) {
				<-time.After(time.Millisecond * 5)
				callback(me, nil)
			}(e.Data)
		},
	}
	go func() {
		_ = r.Start(pctx)
	}()
	defer func() {
		_ = r.Close()
	}()

	r.AddHandler("test-1", rs, 1)
	defer r.RemoveHandler("test-1")

	rs_1 := &ReactiveServiceImpl{
		SelectLogic: func(e Event[mockEventData]) bool {
			return e.Data.name == "errored"
		},
		HandleLogic: func(e Event[mockEventData], callback func(mockEventData, error)) {
			e.Data.Count++
			go func(me mockEventData) {
				<-time.After(time.Millisecond * 5)
				callback(me, errors.New("test-error"))
			}(e.Data)
		},
	}
	r.AddHandler("test-err", rs_1, 1)
	defer r.RemoveHandler("test-err")
	callbackCounter := atomic.Int32{}
	cs := &CallbackServiceImpl{
		SelectLogic: func(c Event[mockEventData]) bool {
			return len(c.Data.name) > 0
		},
		HandleLogic: func(c Event[mockEventData]) {
			if len(c.Data.name) > 0 {
				require.Greater(t, c.Data.Count, int32(0))
			}
			t.Logf("got event %+v", c.Data)
			fmt.Printf("got event %+v", c.Data)
			callbackCounter.Add(1)
		},
	}

	r.AddCallback("test-all", cs, 1)
	defer r.RemoveCallback("test-all")

	n := 4
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5000)
	defer cancel()
	go func(n int) {
		n = n - 2 // because we are doing additional enqueues after this loop
		for n > 0 {
			r.Enqueue(mockEventData{
				name: fmt.Sprintf("test-event-%d", n),
			})
			n--
		}
		r.Enqueue(mockEventData{
			name: "errored",
		})
		res, err := r.EnqueueWait(ctx, mockEventData{
			name: fmt.Sprintf("test-event-%d", n),
		})
		require.NoError(t, err)
		require.Greater(t, res.Count, int32(0))
	}(n)

	for callbackCounter.Load() < int32(n) && ctx.Err() == nil {
		time.Sleep(time.Millisecond * 10)
	}
	<-time.After(time.Second) // to ensure there are no leftovers
	require.Equal(t, callbackCounter.Load(), int32(n))
	require.NoError(t, r.Close())
}

func TestReactor_EventID(t *testing.T) {
	r := New[[]byte, []byte]().(*reactor[[]byte, []byte])

	id := r.genID([]byte("hello-test"))
	require.Len(t, id, 18)
	require.Equal(t, id, IDFromString(id.String()), "id encoding failed")
	require.Equal(t, ID{}, IDFromString(""), "empty id encoding failed")
	require.Equal(t, ID(nil), IDFromString("`^"), "invalid id encoding failed")
}
