package eventloop

import (
	"bytes"
	"context"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEventLoop_Sanity(t *testing.T) {
	el := New[[]byte](1, nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*250)
	defer cancel()

	t.Log("starting event loop")

	go func() {
		_ = el.Start(ctx)
	}()

	helloCount := atomic.Int32{}
	helloF := func(v []byte) {
		helloCount.Add(1)
	}
	el.Register("test", func(v []byte) bool {
		return bytes.Equal(v, []byte("hello"))
	}, 0, helloF, helloF)

	worldCount := atomic.Int32{}
	worldF := func(v []byte) {
		worldCount.Add(1)
	}
	el.Register("test-concurrent", func(v []byte) bool {
		return bytes.Equal(v, []byte("world"))
	}, 2, worldF, worldF, worldF, worldF, worldF)

	el.Register("done", func(v []byte) bool {
		return bytes.Equal(v, []byte("done"))
	}, 0, func(b []byte) {
		t.Log("closing event loop")
		require.NoError(t, el.Close())
	})

	el.Enqueue([]byte("hello"))
	el.Enqueue([]byte("world"))

	for int32(5) > worldCount.Load() && ctx.Err() == nil {
		runtime.Gosched()
	}
	go el.Enqueue([]byte("done"))

	<-ctx.Done()

	require.Equal(t, int32(2), helloCount.Load())
	require.Equal(t, int32(5), worldCount.Load())
}
