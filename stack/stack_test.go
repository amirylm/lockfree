package stack

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/amirylm/lockfree/common"
	"github.com/stretchr/testify/require"
)

func TestStack_Sanity_Int(t *testing.T) {
	n := 32
	common.SanityTest(t, n, New[int], func(i int) int {
		return i + 1
	}, func(i, v int) bool {
		return v == n-i
	})
}

func TestStack_Concurrency_Bytes(t *testing.T) {
	pctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	nmsgs := 1024
	c := 128
	w, r := 2, 2

	reads, writes := common.ConcurrencyTest(t, pctx, c, nmsgs, r, w, New[[]byte], func(i int) []byte {
		return append([]byte{1, 1}, big.NewInt(int64(i)).Bytes()...)
	}, func(i int, v []byte) bool {
		return len(v) > 1 && v[0] == 1
	})

	expectedW := int64(nmsgs * w) // num of msgs * num of writers
	require.Equal(t, expectedW, writes, "num of writes is wrong")
	expectedR := int64(nmsgs * r) // num of msgs * num of writers
	require.Equal(t, expectedR, reads, "num of reads is wrong")
}

func TestStack_Range(t *testing.T) {
	nitems := 10
	s := New[int](nitems * 2).(*Stack[int])
	for i := 0; i < nitems; i++ {
		require.True(t, s.Push(i+1))
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
