package benchmark

import (
	"testing"

	"github.com/amirylm/lockfree/stack"
)

func BenchmarkStack(b *testing.B) {
	s := stack.New[int]()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Push(i)
		_, _ = s.Pop()
	}
}
