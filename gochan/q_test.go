package gochan

import (
	"context"
	"testing"
	"time"

	"github.com/amirylm/lockfree/common"
	"github.com/stretchr/testify/require"
)

func TestGoChanQ_EnqueueDequeue(t *testing.T) {
	nitems := 10
	q := New[int](nitems)
	for i := 0; i < nitems; i++ {
		d := i + 1
		require.NoError(t, q.Push(d))
	}
	require.Error(t, q.Push(11))
	require.Equal(t, nitems, q.Size())
	for i := 0; i < nitems; i++ {
		val, ok := q.Pop()
		require.True(t, ok, i)
		require.NotNil(t, val, i)
		require.Equal(t, i+1, val)
		require.Equal(t, nitems-(i+1), q.Size())
	}
	require.Equal(t, 0, q.Size())
	require.NoError(t, q.Push(1))
}

func TestGoChanQ_Concurrency(t *testing.T) {
	pctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*2))
	defer cancel()

	common.DoConcurrencyCheck(pctx, t, New[[]byte](128), 100, 1, 1)
}
