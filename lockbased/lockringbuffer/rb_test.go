package lockringbuffer

import (
	"context"
	"testing"
	"time"

	"github.com/amirylm/lockfree/common"
)

func TestLockRingBuffer_Sanity_Int(t *testing.T) {
	common.SanityTest(t, 32, New[int], func(i int) int {
		return i + 1
	}, func(i, v int) bool {
		return v == i+1
	})
}

func TestLockRingBuffer_Concurrency_Bytes(t *testing.T) {
	pctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	common.ConcurrencyTest(t, pctx, 128, 1024, 2, 2, New[[]byte], func(i int) []byte {
		return []byte{1, 1, 1, 1}
	}, func(i int, v []byte) bool {
		return len(v) == 4 && v[0] == 1
	})
}
