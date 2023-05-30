package benchmark

import (
	"testing"

	"github.com/amirylm/lockfree/common"
	"github.com/amirylm/lockfree/queue"
	"github.com/amirylm/lockfree/ringbuffer"
	"github.com/amirylm/lockfree/stack"
)

func BenchmarkInt_PushPop(b *testing.B) {
	tests := []struct {
		name       string
		collection common.LockFreeCollection[int]
	}{
		{
			"ring buffer",
			ringbuffer.New[int](128),
		},
		{
			"stack",
			stack.New[int](128),
		},
		{
			"queue",
			queue.New[int](128),
		},
	}

	for _, tc := range tests {
		collection := tc.collection
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = collection.Push(i)
				_, _ = collection.Pop()
			}
		})
	}
}
