package rb_lock

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/amirylm/lockfree/utils"
	"github.com/stretchr/testify/require"
)

func TestRingBufferLock_Sanity_Int(t *testing.T) {
	utils.SanityTest(t, 32, New[int], func(i int) int {
		return i + 1
	}, func(i, v int) bool {
		return v == i+1
	})
}

func TestRingBufferLock_Concurrency_Bytes(t *testing.T) {
	pctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	nmsgs := 1024
	c := 128
	w, r := 2, 2

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
