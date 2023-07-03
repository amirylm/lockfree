package ringbuffer

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/amirylm/lockfree/utils"
	"github.com/stretchr/testify/require"
)

func TestRingBuffer_Sanity_Int(t *testing.T) {
	utils.SanityTest(t, 32, New[int], func(i int) int {
		return i + 1
	}, func(i, v int) bool {
		return v == i+1
	})
}

func TestRingBuffer_Concurrency_Bytes(t *testing.T) {
	pctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	nmsgs := 1024
	c := 128
	w, r := 5, 5

	reads, writes := utils.ConcurrencyTest(t, pctx, c, nmsgs, r, w, New[[]byte], func(i int) []byte {
		return append([]byte{1, 1}, big.NewInt(int64(i)).Bytes()...)
	}, func(i int, v []byte) bool {
		return len(v) > 1 && v[0] == 1
	})

	expectedW := int64(nmsgs * w) // num of msgs * num of writers
	require.Equal(t, expectedW, writes, "num of writes is wrong")
	expectedR := int64(nmsgs * r) // num of msgs * num of writers
	require.Equal(t, expectedR, reads, "num of reads is wrong")
}

func TestRingBuffer_Overflow(t *testing.T) {
	rb := NewWithOverride[int](128)
	overflow := 5
	n := 128
	require.True(t, rb.Empty(), "should be empty")
	for i := 0; i < n+overflow; i++ {
		require.True(t, rb.Enqueue(i+1), "failed to enqueue element in index %d", i)
	}
	for i := 0; i < n; i++ {
		v, ok := rb.Dequeue()
		require.True(t, ok, "failed to read element in index %d", i)
		require.Equal(t, i+overflow+1, v)
	}
}
