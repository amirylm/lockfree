package benchmark

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/amirylm/lockfree/common"
	"github.com/amirylm/lockfree/queue"

	"github.com/amirylm/lockfree/gochan"
	"github.com/amirylm/lockfree/lockringbuffer"
	"github.com/amirylm/lockfree/ringbuffer"
	"github.com/amirylm/lockfree/stack"
)

func BenchmarkInt_PushPop_Single(b *testing.B) {
	tests := []struct {
		name       string
		collection common.DataStructure[int]
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
		{
			"go channel",
			gochan.New[int](128),
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

func BenchmarkInt_PushPop_RW(b *testing.B) {
	tests := []concurrentTestCase[int]{
		{
			"ring buffer (atomic)",
			ringbuffer.New[int](128),
			1,
			1,
		},
		{
			"ring buffer (lock)",
			lockringbuffer.New[int](128),
			1,
			1,
		},
		{
			"stack",
			stack.New[int](128),
			1,
			1,
		},
		{
			"queue",
			queue.New[int](128),
			1,
			1,
		},
		{
			"go channel",
			gochan.New[int](128),
			1,
			1,
		},
	}

	for _, tc := range tests {
		b.Run(testName(tc.name, tc.readers, tc.writers), testCaseInt(tc, b))
	}
}

func BenchmarkInt_PushPopConcurrentLoad(b *testing.B) {
	tests := []concurrentTestCase[int]{
		{
			"ring buffer (atomic)",
			ringbuffer.New[int](128),
			3,
			2,
		},
		{
			"ring buffer (lock)",
			lockringbuffer.New[int](128),
			3,
			2,
		},
		{
			"go channel",
			gochan.New[int](128),
			3,
			2,
		},
		{
			"ring buffer (atomic)",
			ringbuffer.New[int](128),
			2,
			3,
		},
		{
			"ring buffer (lock)",
			lockringbuffer.New[int](128),
			2,
			3,
		},
		{
			"go channel",
			gochan.New[int](128),
			2,
			3,
		},
		{
			"ring buffer (atomic)",
			ringbuffer.New[int](128),
			5,
			5,
		},
		{
			"ring buffer (lock)",
			lockringbuffer.New[int](128),
			5,
			5,
		},
		{
			"go channel",
			gochan.New[int](128),
			5,
			5,
		},
		{
			"ring buffer (atomic)",
			ringbuffer.New[int](128),
			15,
			2,
		},
		{
			"ring buffer (lock)",
			lockringbuffer.New[int](128),
			15,
			2,
		},
		{
			"go channel 15",
			gochan.New[int](128),
			15,
			2,
		},
		{
			"ring buffer (atomic)",
			ringbuffer.New[int](128),
			15,
			15,
		},
		{
			"ring buffer (lock)",
			lockringbuffer.New[int](128),
			15,
			15,
		},
		{
			"go channel",
			gochan.New[int](128),
			15,
			15,
		},
		{
			"ring buffer (atomic)",
			ringbuffer.New[int](128),
			2,
			15,
		},
		{
			"ring buffer (lock)",
			lockringbuffer.New[int](128),
			2,
			15,
		},
		{
			"go channel",
			gochan.New[int](128),
			2,
			15,
		},
	}

	for _, tc := range tests {
		b.Run(testName(tc.name, tc.readers, tc.writers), testCaseInt(tc, b))
	}
}

func testName(name string, r, w int) string {
	return fmt.Sprintf("%s %d R %d W", name, r, w)
}

type concurrentTestCase[V any] struct {
	name    string
	ds      common.DataStructure[V]
	readers int
	writers int
}

func testCaseInt(tc concurrentTestCase[int], b *testing.B) func(b *testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		collection := tc.ds
		for i := 0; i < b.N; i++ {
			n := tc.writers
			for n > 0 {
				n--
				go func(i int) {
					for !collection.Push(i) {
						runtime.Gosched()
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
	}
}
