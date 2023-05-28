package ringbuffer

import (
	"context"
	"testing"
	"time"

	"github.com/amirylm/lockfree/common"
	"github.com/stretchr/testify/require"
)

func TestRingBuffer_EnqueueDequeue(t *testing.T) {
	nitems := 10
	rb := New[int](nitems)
	for i := 0; i < nitems; i++ {
		d := i + 1
		require.NoError(t, rb.Enqueue(d))
	}
	require.Error(t, rb.Enqueue(11))
	require.Equal(t, nitems, rb.Len())
	for i := 0; i < nitems; i++ {
		val, ok := rb.Dequeue()
		require.NoError(t, ok, i)
		require.NotNil(t, val, i)
		require.Equal(t, i+1, val)
		require.Equal(t, nitems-(i+1), rb.Len())
	}
	require.Equal(t, 0, rb.Len())
	require.NoError(t, rb.Enqueue(1))
}

func TestRingBuffer_Concurrency(t *testing.T) {
	pctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*2))
	defer cancel()

	common.DoConcurrencyCheck(pctx, t, &wrapper[[]byte]{rb: New[[]byte](128)}, 100, 1, 1)
}

type wrapper[Value any] struct {
	rb *RingBuffer[Value]
}

func (w *wrapper[Value]) Store(v Value) error {
	return w.rb.Enqueue(v)
}

func (w *wrapper[Value]) Read() (*Value, error) {
	v, err := w.rb.Dequeue()
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (w *wrapper[Value]) Len() int {
	return w.rb.Len()
}

func (w *wrapper[Value]) IsEmpty() bool {
	return w.rb.IsEmpty()
}
