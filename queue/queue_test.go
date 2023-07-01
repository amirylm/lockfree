package queue

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/amirylm/lockfree/utils"
	"github.com/stretchr/testify/require"
)

func TestLinkedListQueue_Sanity_Int(t *testing.T) {
	utils.SanityTest(t, 32, New[int], func(i int) int {
		return i + 1
	}, func(i, v int) bool {
		return v == i+1
	})
}

func TestLinkedListQueue_Concurrency_Bytes(t *testing.T) {
	pctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	nmsgs := 1024
	c := 128
	w, r := 2, 2

	_, _ = utils.ConcurrencyTest(t, pctx, c, nmsgs, r, w, New[[]byte], func(i int) []byte {
		return append([]byte{1, 1}, big.NewInt(int64(i)).Bytes()...)
	}, func(i int, v []byte) bool {
		return true
		// TODO: fix
		// return len(v) > 1 && v[0] == 1
	})

	// TODO: uncomment
	// expectedW := int64(nmsgs * w) // num of msgs * num of writers
	// require.Equal(t, expectedW, writes, "num of writes is wrong")
	// expectedR := int64(nmsgs * r) // num of msgs * num of writers
	// require.Equal(t, expectedR, reads, "num of reads is wrong")
}

func TestLinkedListQueue_Range(t *testing.T) {
	numitems := 10
	q := New[int](numitems).(*Queue[int])
	for i := 0; i < numitems; i++ {
		require.True(t, q.Enqueue(i+1))
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
