package benchmark

import (
	"testing"

	"github.com/amirylm/lockfree/ringbuffer"
)

func BenchmarkRingBuffer(b *testing.B) {
	q := ringbuffer.New[int](128)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Enqueue(i)
		_, _ = q.Dequeue()
	}
}
