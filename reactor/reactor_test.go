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

func TestReactor(t *testing.T) {
	r := New[[]byte]()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	go func() {
		t.Log("starting reactor")
		_ = r.Start(ctx)
	}()

	t.Run("sanity", func(t *testing.T) {
		count, f := newCounter[[]byte]()
		expectedCount := int32(3)
		r.Register("test", func(v []byte) bool {
			return bytes.Equal(v, []byte("hello"))
		}, 0, f, f, f)
		defer r.Unregister("test")

		// double register should not work
		r.Register("test", func(v []byte) bool {
			return bytes.Equal(v, []byte("hello"))
		}, 0, f)

		go r.Enqueue([]byte("hello"))
		go r.Enqueue([]byte("world"))
		go r.Enqueue([]byte("hello-world"))

		for count.Load() < expectedCount && ctx.Err() == nil {
			runtime.Gosched()
		}
		require.Equal(t, expectedCount, count.Load())
	})

	t.Run("workers", func(t *testing.T) {
		count, f := newCounter[[]byte]()
		expectedCount := int32(8 * 2)
		r.Register("test-workers", func(v []byte) bool {
			return bytes.Equal(v, []byte("hello")) || bytes.Equal(v, []byte("world"))
		}, 2, f, f, f, f, f, f, f, f)
		defer r.Unregister("test-workers")

		count2, f2 := newCounter[[]byte]()
		expectedCount2 := int32(4)
		r.Register("test-workers-2", func(v []byte) bool {
			return bytes.Equal(v, []byte("hello"))
		}, 2, f2, f2, f2, f2)
		defer r.Unregister("test-workers-2")

		go r.Enqueue([]byte("hello"))
		go r.Enqueue([]byte("world"))
		go r.Enqueue([]byte("hello-world"))

		for count.Load() < expectedCount && ctx.Err() == nil {
			runtime.Gosched()
		}
		require.Equal(t, expectedCount, count.Load())
		require.Equal(t, expectedCount2, count2.Load())
	})

	t.Run("close", func(t *testing.T) {
		require.NoError(t, r.Close())
		require.NoError(t, r.Close())
		r.Enqueue([]byte("hello"))
	})
}

func TestReactor_handleControl(t *testing.T) {
	r := New(
		WithEventQueue(ringbuffer.New(ringbuffer.WithCapacity[[]byte](32))),
		WithControlQueue(ringbuffer.New(ringbuffer.WithCapacity[controlEvent[[]byte]](32))),
	)

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
				services = r.handleControl(services, e)
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
