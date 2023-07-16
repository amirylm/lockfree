package reactor

import (
	"bytes"
	"context"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amirylm/lockfree/ringbuffer"
	"github.com/stretchr/testify/require"
)

func TestDemux(t *testing.T) {
	d := NewDemux(WithCloneFn(func(b []byte) []byte {
		if len(b) == 0 {
			return b
		}
		cp := make([]byte, len(b))
		copy(b, cp)
		return cp
	})).(*demultiplexer[[]byte])
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	go func() {
		t.Log("starting reactor")
		_ = d.Start(ctx)
	}()

	t.Run("sanity", func(t *testing.T) {
		count, f := newCounter[[]byte]()
		expectedCount := int32(3)
		d.Register("test", func(v []byte) bool {
			return bytes.Equal(v, []byte("hello"))
		}, 0, f, f, f)
		defer d.Unregister("test")

		// double register should not work
		d.Register("test", func(v []byte) bool {
			return bytes.Equal(v, []byte("hello"))
		}, 0, f)

		go d.Enqueue([]byte("hello"))
		go d.Enqueue([]byte("world"))
		go d.Enqueue([]byte("hello-world"))

		for count.Load() < expectedCount && ctx.Err() == nil {
			runtime.Gosched()
		}
		require.Equal(t, expectedCount, count.Load())
	})

	t.Run("workers", func(t *testing.T) {
		count, f := newCounter[[]byte]()
		expectedCount := int32(8 * 2)
		d.Register("test-workers", func(v []byte) bool {
			return bytes.Equal(v, []byte("hello")) || bytes.Equal(v, []byte("world"))
		}, 2, f, f, f, f, f, f, f, f)
		defer d.Unregister("test-workers")

		count2, f2 := newCounter[[]byte]()
		expectedCount2 := int32(4)
		d.Register("test-workers-2", func(v []byte) bool {
			return bytes.Equal(v, []byte("hello"))
		}, 2, f2, f2, f2, f2)
		defer d.Unregister("test-workers-2")

		go d.Enqueue([]byte("hello"))
		go d.Enqueue([]byte("world"))
		go d.Enqueue([]byte("hello-world"))

		for count.Load() < expectedCount && ctx.Err() == nil {
			runtime.Gosched()
		}
		require.Equal(t, expectedCount, count.Load())
		require.Equal(t, expectedCount2, count2.Load())
	})

	t.Run("register no handlers", func(t *testing.T) {
		d.Register("test-workers-2", func(v []byte) bool {
			return true
		}, 0)
		d.Enqueue([]byte("hello"))
	})

	t.Run("close", func(t *testing.T) {
		require.NoError(t, d.Close())
		require.NoError(t, d.Close())
		d.Enqueue([]byte("hello"))
	})
}

func TestDemux_handleControl(t *testing.T) {
	r := NewDemux(
		WithEventQueue(ringbuffer.New(ringbuffer.WithCapacity[[]byte](32))),
		WithControlQueue(ringbuffer.New(ringbuffer.WithCapacity[controlEvent[[]byte]](32))),
	).(*demultiplexer[[]byte])

	tests := []struct {
		name     string
		existing []service[[]byte]
		events   []controlEvent[[]byte]
		want     []service[[]byte]
	}{
		{
			name:     "empty",
			existing: nil,
			events: []controlEvent[[]byte]{
				{
					control: registerService,
					svc: service[[]byte]{
						id: "test",
					},
				},
			},
			want: []service[[]byte]{
				{
					id: "test",
				},
			},
		},
		{
			name: "unknown control",
			existing: []service[[]byte]{
				{
					id: "test",
				},
			},
			events: []controlEvent[[]byte]{
				{
					control: 5,
					svc: service[[]byte]{
						id: "test-2",
					},
				},
			},
			want: []service[[]byte]{
				{
					id: "test",
				},
			},
		},
		{
			name: "register",
			existing: []service[[]byte]{
				{
					id: "test",
				},
			},
			events: []controlEvent[[]byte]{
				{
					control: registerService,
					svc: service[[]byte]{
						id: "test2",
					},
				},
			},
			want: []service[[]byte]{
				{
					id: "test",
				},
				{
					id: "test2",
				},
			},
		},
		{
			name: "double register",
			existing: []service[[]byte]{
				{
					id: "test",
				},
			},
			events: []controlEvent[[]byte]{
				{
					control: registerService,
					svc: service[[]byte]{
						id: "test",
					},
				},
			},
			want: []service[[]byte]{
				{
					id: "test",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			services := tc.existing
			for _, e := range tc.events {
				services = r.handleControl(services, &e)
			}
			require.Equal(t, tc.want, services)
		})
	}
}

func newCounter[T any]() (*atomic.Int32, func(T)) {
	count := atomic.Int32{}
	return &count, func(T) {
		count.Add(1)
	}
}
