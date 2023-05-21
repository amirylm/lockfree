package stack

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStack_PushPop(t *testing.T) {
	nitems := 10
	s := New[int]()
	for i := 0; i < nitems; i++ {
		s.Push(i + 1)
	}
	require.Equal(t, nitems, s.Len())
	for i := 0; i < nitems; i++ {
		val, ok := s.Pop()
		require.True(t, ok)
		require.Equal(t, nitems-(i+1)+1, *val)
		require.Equal(t, nitems-(i+1), s.Len())
	}
	require.Equal(t, 0, s.Len())
}

func TestQueue_Range(t *testing.T) {
	nitems := 10
	s := New[int]()
	for i := 0; i < nitems; i++ {
		s.Push(i + 1)
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
