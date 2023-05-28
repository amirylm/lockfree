package stack

import (
	"context"
	"errors"
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
	require.True(t, s.IsFull())
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

	common.DoConcurrencyCheck(pctx, t, &wrapper[[]byte]{stack: New[[]byte](128)}, 100, 1, 1)
}

type wrapper[Value any] struct {
	stack *Stack[Value]
}

func (w *wrapper[Value]) Store(v Value) error {
	return w.stack.Push(v)
}

func (w *wrapper[Value]) Read() (*Value, error) {
	v, ok := w.stack.Pop()
	if !ok {
		return nil, errors.New("err")
	}
	return v, nil
}

func (w *wrapper[Value]) Len() int {
	return w.stack.Len()
}

func (w *wrapper[Value]) IsEmpty() bool {
	return w.stack.Len() == 0
}
