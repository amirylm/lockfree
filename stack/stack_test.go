package stack

import (
	"context"
	"testing"
	"time"

	"github.com/amirylm/lockfree/common"
	"github.com/stretchr/testify/require"
)

func TestStack_PushPop(t *testing.T) {
	nitems := 10
	s := New[int](nitems)
	for i := 0; i < nitems; i++ {
		require.NoError(t, s.Push(i+1))
	}
	require.Error(t, s.Push(11))
	require.True(t, s.Full())
	require.Equal(t, nitems, s.Size())
	for i := 0; i < nitems; i++ {
		val, ok := s.Pop()
		require.True(t, ok)
		require.Equal(t, nitems-(i+1)+1, val)
		require.Equal(t, nitems-(i+1), s.Size())
	}
	require.Equal(t, 0, s.Size())
}

func TestQueue_Range(t *testing.T) {
	nitems := 10
	s := New[int](nitems * 2)
	for i := 0; i < nitems; i++ {
		require.NoError(t, s.Push(i+1))
	}

	tests := []struct {
		name     string
		stack    *Stack[int]
		iterator func(val int) bool
		want     int
	}{
		{
			name:  "happy flow",
			stack: s,
			iterator: func(val int) bool {
				return false
			},
			want: nitems,
		},
		{
			name:  "stop half way",
			stack: s,
			iterator: func(val int) bool {
				return val == nitems/2
			},
			want: nitems/2 + 1,
		},
		{
			name:  "stop",
			stack: s,
			iterator: func(val int) bool {
				return true
			},
			want: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			count := 0
			tc.stack.Range(func(val int) bool {
				count++
				return tc.iterator(val)
			})
			require.Equal(t, tc.want, count)
		})
	}
}

func TestStack_Concurrency(t *testing.T) {
	pctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*2))
	defer cancel()

	common.DoConcurrencyCheck(pctx, t, New[[]byte](128), 100, 1, 1)
}
