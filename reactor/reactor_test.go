package reactor

import (
	"bytes"
	"context"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReactor(t *testing.T) {
	r := New[[]byte](nil)
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
		r.Enqueue([]byte("hello"))
	})
}

func newCounter[T any]() (*atomic.Int32, func(T)) {
	count := atomic.Int32{}
	return &count, func(T) {
		count.Add(1)
	}
}
