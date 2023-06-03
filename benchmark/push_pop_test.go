package benchmark

import (
	"testing"

	"github.com/amirylm/lockfree/common"
	"github.com/amirylm/lockfree/queue"

	"github.com/amirylm/lockfree/gochan"
	"github.com/amirylm/lockfree/lockringbuffer"
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
			"ring buffer (lock)",
			lockringbuffer.New[int](128),
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

func BenchmarkInt_PushPopConcurrent(b *testing.B) {
	tests := []struct {
		name       string
		collection common.LockFreeCollection[int]
		readers    int
		writers    int
	}{
		{
			"ring buffer 1 R 1 W",
			ringbuffer.New[int](128),
			1,
			1,
		},
		{
			"ring buffer (lock) 1 R 1 W",
			lockringbuffer.New[int](128),
			1,
			1,
		},
		{
			"stack 1 R 1 W",
			stack.New[int](128),
			1,
			1,
		},
		{
			"queue 1 R 1 W",
			queue.New[int](128),
			1,
			1,
		},
		{
			"go channel 1 R 1 W",
			gochan.New[int](128),
			1,
			1,
		},
	}

	for _, tc := range tests {
		collection := tc.collection
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				n := tc.writers
				for n > 0 {
					n--
					go func(i int) {
						for collection.Push(i) != nil {
						}
					}(i)
				}
				n = tc.readers
				for n > 0 {
					n--
					go func(i, n int) {
						_, _ = collection.Pop()
					}(i, tc.readers)
				}
			}
		})
	}
}

func BenchmarkInt_PushPopConcurrentLoad(b *testing.B) {
	tests := []struct {
		name       string
		collection common.LockFreeCollection[int]
		readers    int
		writers    int
	}{
		{
			"ring buffer (lock) 3 R 2 W",
			lockringbuffer.New[int](128),
			3,
			2,
		},
		{
			"go channel 3 R 2 W",
			gochan.New[int](128),
			3,
			2,
		},
		{
			"ring buffer (lock) 2 R 3 W",
			lockringbuffer.New[int](128),
			2,
			3,
		},
		{
			"go channel 2 R 3 W",
			gochan.New[int](128),
			2,
			3,
		},
		{
			"ring buffer (lock) 5 R 5 W",
			lockringbuffer.New[int](128),
			5,
			5,
		},
		{
			"go channel 5 R 5 W",
			gochan.New[int](128),
			5,
			5,
		},
		{
			"ring buffer (lock) 15 R 2 W",
			lockringbuffer.New[int](128),
			15,
			2,
		},
		{
			"go channel 15 R 2 W",
			gochan.New[int](128),
			15,
			2,
		},
		{
			"ring buffer (lock) 15 R 15 W",
			lockringbuffer.New[int](128),
			15,
			15,
		},
		{
			"go channel 15 R 15 W",
			gochan.New[int](128),
			15,
			15,
		},
		{
			"ring buffer (lock) 2 R 15 W",
			lockringbuffer.New[int](128),
			2,
			15,
		},
		{
			"go channel 2 R 15 W",
			gochan.New[int](128),
			2,
			15,
		},
	}

	for _, tc := range tests {
		collection := tc.collection
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				n := tc.writers
				for n > 0 {
					n--
					go func(i int) {
						_ = collection.Push(i)
					}(i)
				}
				n = tc.readers
				for n > 0 {
					n--
					go func(i, n int) {
						_, _ = collection.Pop()
					}(i, tc.readers)
				}
			}
		})
	}
}
