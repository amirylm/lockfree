package queue

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueue_PushPop(t *testing.T) {
	numitems := 10
	q := New[int](int32(numitems))
	for i := 0; i < numitems; i++ {
		q.Push(i + 1)
	}
	require.Equal(t, numitems, q.Size())
	for i := 0; i < numitems; i++ {
		val, ok := q.Pop()
		require.True(t, ok)
		require.Equal(t, (i + 1), val)
		require.Equal(t, numitems-(i+1), q.Size())
	}
	require.Equal(t, 0, q.Size())
}

func TestQueue_Range(t *testing.T) {
	numitems := 10
	q := New[int](int32(numitems))
	for i := 0; i < numitems; i++ {
		q.Push(i + 1)
	}

	tests := []struct {
		name     string
		queue    *Queue[int]
		iterator func(val int) bool
		want     int
	}{
		{
			name:  "happy flow",
			queue: q,
			iterator: func(val int) bool {
				return false
			},
			want: numitems,
		},
		{
			name:  "stop half way",
			queue: q,
			iterator: func(val int) bool {
				return val == numitems/2
			},
			want: numitems / 2,
		},
		{
			name:  "stop",
			queue: q,
			iterator: func(val int) bool {
				return true
			},
			want: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			count := 0
			tc.queue.Range(func(val int) bool {
				count++
				return tc.iterator(val)
			})
			require.Equal(t, tc.want, count)
		})
	}
}
