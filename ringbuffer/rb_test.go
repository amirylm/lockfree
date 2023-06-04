package ringbuffer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/amirylm/lockfree/common"
	"github.com/stretchr/testify/require"
)

func TestRingBuffer_EnqueueDequeue(t *testing.T) {
	nitems := 10
	rb := New[string](nitems)
	for i := 0; i < nitems; i++ {
		d := i + 1
		require.NoError(t, rb.Push(fmt.Sprintf("%d", d)))
	}
	require.Error(t, rb.Push("11"))
	require.Equal(t, nitems, rb.Size())
	for i := 0; i < nitems; i++ {
		val, ok := rb.Pop()
		require.True(t, ok, i)
		require.Equal(t, fmt.Sprintf("%d", i+1), val)
		require.Equal(t, nitems-(i+1), rb.Size())
	}
	require.Equal(t, 0, rb.Size())
	require.NoError(t, rb.Push("11"))
}

func TestRingBuffer_Concurrency(t *testing.T) {
	pctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*2))
	defer cancel()

	common.DoConcurrencyCheck(pctx, t, New[[]byte](128), 100, 5, 5)
}
