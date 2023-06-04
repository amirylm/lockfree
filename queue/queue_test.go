package queue

import (
	"context"
	"testing"
	"time"

	"github.com/amirylm/lockfree/common"
	"github.com/stretchr/testify/require"
)

func TestLinkedListQueue_Sanity_Int(t *testing.T) {
	common.SanityTest(t, 32, func(capacity int) common.DataStructure[int] {
		return New[int](capacity)
	}, func(i int) int {
		return i + 1
	}, func(i, v int) bool {
		return v == i+1
	})
}

func TestLinkedListQueue_Concurrency_Bytes(t *testing.T) {
	pctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	common.ConcurrencyTest(t, pctx, 128, 128, 1, 1, func(capacity int) common.DataStructure[[]byte] {
		return New[[]byte](capacity)
	}, func(i int) []byte {
		return []byte{1, 1, 1, 1}
	}, func(i int, v []byte) bool {
		return len(v) == 4 && v[0] == 1
	})
}

func TestLinkedListQueue_Range(t *testing.T) {
	numitems := 10
	q := New[int](numitems)
	for i := 0; i < numitems; i++ {
		require.True(t, q.Push(i+1))
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
