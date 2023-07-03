package eventloop

import (
	"bytes"
	"context"
	"testing"
)

func TestEventLoop_Sanity(t *testing.T) {
	el := New[[]byte](1, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Log("starting event loop")

	go func() {
		_ = el.Start(ctx)
	}()

	logf := func(v []byte) {
		t.Logf("got: %s", v)
	}

	el.Register("test", func(v []byte) bool {
		return bytes.Equal(v, []byte("hello"))
	}, 0, logf, logf)

	el.Register("test-concurrent", func(v []byte) bool {
		return bytes.Equal(v, []byte("world"))
	}, 2, logf, logf, logf)

	el.Register("done", func(v []byte) bool {
		return bytes.Equal(v, []byte("done"))
	}, 0, func(b []byte) {
		cancel()
	})

	el.Enqueue([]byte("hello"))
	el.Enqueue([]byte("world"))

	go el.Enqueue([]byte("done"))

	<-ctx.Done()
}
