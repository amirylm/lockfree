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
		require.NoError(t, rb.Push(d))
	}
	require.Error(t, rb.Push(11))
	require.Equal(t, nitems, rb.Size())
	for i := 0; i < nitems; i++ {
		val, ok := rb.Pop()
		require.True(t, ok, i)
		require.NotNil(t, val, i)
		require.Equal(t, i+1, val)
		require.Equal(t, nitems-(i+1), rb.Size())
	}
	require.Equal(t, 0, rb.Size())
	require.NoError(t, rb.Push(1))
}

func TestRingBuffer_Concurrency(t *testing.T) {
	pctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*2))
	defer cancel()

	common.DoConcurrencyCheck(pctx, t, New[[]byte](128), 100, 1, 1)
}
