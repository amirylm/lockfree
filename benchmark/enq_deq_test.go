package benchmark

import (
	"fmt"
	"testing"

	"github.com/amirylm/lockfree/core"
	"github.com/amirylm/lockfree/queue"

	"github.com/amirylm/lockfree/benchmark/gochan"
	"github.com/amirylm/lockfree/benchmark/rb_lock"
	"github.com/amirylm/lockfree/ringbuffer"
	"github.com/amirylm/lockfree/stack"
)

func Benchmark_EnqueueDequeueInt(b *testing.B) {
	tests := []struct {
		name       string
		collection core.Queue[int]
	}{
		{
			"ring buffer",
			ringbuffer.New[int](128),
		},
		{
			"ring buffer (lock)",
			rb_lock.New[int](128),
		},
		{
			"stack",
			stack.NewQueueAdapter[int](128),
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
				_ = collection.Enqueue(i)
				_, _ = collection.Dequeue()
			}
		})
	}
}

func Benchmark_PushPopBytes_Concurrent_Single(b *testing.B) {
	benchmarkPushPopBytes(b, 128, 1, 1)
}

func Benchmark_PushPopBytes_Concurrent_Multi_4(b *testing.B) {
	benchmarkPushPopBytes(b, 128, 4, 4)
}

func Benchmark_PushPopBytes_Concurrent_Multi_16(b *testing.B) {
	benchmarkPushPopBytes(b, 128, 16, 16)
}

func Benchmark_PushPopBytes_Concurrent_Multi_Readers(b *testing.B) {
	benchmarkPushPopBytes(b, 128, 4, 2)
}

func Benchmark_PushPopBytes_Concurrent_Multi_Readers_16(b *testing.B) {
	benchmarkPushPopBytes(b, 128, 16, 2)
}

func Benchmark_PushPopBytes_Concurrent_Multi_Readers_64(b *testing.B) {
	benchmarkPushPopBytes(b, 128, 64, 4)
}

func Benchmark_PushPopBytes_Concurrent_Multi_Writers(b *testing.B) {
	benchmarkPushPopBytes(b, 128, 2, 4)
}

func Benchmark_PushPopBytes_Concurrent_Multi_Writers_16(b *testing.B) {
	benchmarkPushPopBytes(b, 128, 2, 16)
}

func Benchmark_PushPopBytes_Concurrent_Multi_Writers_64(b *testing.B) {
	benchmarkPushPopBytes(b, 128, 4, 64)
}

func benchmarkPushPopBytes(b *testing.B, c, r, w int) {
	tests := []concurrentTestCase[[]byte]{
		{
			"ring buffer (atomic)",
			ringbuffer.New[[]byte](c),
			r,
			w,
		},
		{
			"ring buffer (lock)",
			rb_lock.New[[]byte](c),
			r,
			w,
		},
		{
			"queue",
			queue.New[[]byte](c),
			r,
			w,
		},
		{
			"go channel",
			gochan.New[[]byte](c),
			r,
			w,
		},
		{
			"stack",
			stack.NewQueueAdapter[[]byte](c),
			r,
			w,
		},
	}

	for _, tc := range tests {
		b.Run(testName(tc.name, tc.readers, tc.writers), testCaseBytes(tc, b))
	}
}

func testName(name string, r, w int) string {
	return fmt.Sprintf("%s %d R %d W", name, r, w)
}

type concurrentTestCase[V any] struct {
	name    string
	ds      core.Queue[V]
	readers int
	writers int
}

// func testCaseInt(tc concurrentTestCase[int], b *testing.B) func(b *testing.B) {
// 	return func(b *testing.B) {
// 		b.ReportAllocs()
// 		b.ResetTimer()
// 		collection := tc.ds
// 		for i := 0; i < b.N; i++ {
// 			n := tc.writers
// 			for n > 0 {
// 				n--
// 				go func(i int) {
// 					_ = collection.Enqueue(i)
// 				}(i)
// 			}
// 			n = tc.readers
// 			for n > 0 {
// 				n--
// 				go func(i, n int) {
// 					_, _ = collection.Dequeue()
// 				}(i, tc.readers)
// 			}
// 		}
// 	}
// }

func testCaseBytes(tc concurrentTestCase[[]byte], b *testing.B) func(b *testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		collection := tc.ds
		for i := 0; i < b.N; i++ {
			n := tc.writers
			for n > 0 {
				n--
				go func(i int) {
					_ = collection.Enqueue([]byte(fmt.Sprintf("%06d", i)))
				}(i)
			}
			n = tc.readers
			for n > 0 {
				n--
				go func(i, n int) {
					_, _ = collection.Dequeue()
				}(i, tc.readers)
			}
		}
	}
}
