package benchmark

import (
	"testing"

	"github.com/amirylm/lockfree/queue"
)

func BenchmarkStack(b *testing.B) {
	s := queue.New[int]()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(i)
		_, _ = s.Pop()
	}
}
