package load

import (
	"testing"

	"github.com/amirylm/lockfree/benchmark"
)

func Bench_Concurrent_Multi_Writers_64(b *testing.B) {
	benchmark.BenchEnqueueDequeueBytes(b, 128, 4, 64)
}

func Bench_Concurrent_Multi_Readers_64(b *testing.B) {
	benchmark.BenchEnqueueDequeueBytes(b, 128, 64, 4)
}
