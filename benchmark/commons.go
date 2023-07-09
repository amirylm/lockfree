package benchmark

import (
	"fmt"
	"testing"

	"github.com/amirylm/lockfree/benchmark/gochan"
	"github.com/amirylm/lockfree/benchmark/rb_lock"
	"github.com/amirylm/lockfree/core"
	"github.com/amirylm/lockfree/queue"
	"github.com/amirylm/lockfree/ringbuffer"
	"github.com/amirylm/lockfree/stack"
)

func BenchEnqueueDequeueBytes(b *testing.B, c, r, w int) {
	tests := []concurrentTestCase[[]byte]{
		{
			"ring buffer queue",
			ringbuffer.New[[]byte](ringbuffer.WithCapacity[[]byte](c)),
			r,
			w,
		},
		{
			"ring buffer queue (lock)",
			rb_lock.New[[]byte](rb_lock.WithCapacity[[]byte](c)),
			r,
			w,
		},
		{
			"linked list queue",
			queue.New[[]byte](queue.WithCapacity[[]byte](c)),
			r,
			w,
		},
		{
			"go chan",
			gochan.New[[]byte](gochan.WithCapacity[[]byte](c)),
			r,
			w,
		},
		{
			"linked list stack",
			stack.NewQueueAdapter[[]byte](c),
			r,
			w,
		},
	}

	for _, tc := range tests {
		b.Run(testName(tc.name, tc.readers, tc.writers), testCaseBytes(tc, b))
	}
}

func BenchEnqueueDequeueInt(b *testing.B, c, r, w int) {
	tests := []concurrentTestCase[int]{
		{
			"ring buffer queue",
			ringbuffer.New[int](ringbuffer.WithCapacity[int](c)),
			r,
			w,
		},
		{
			"ring buffer queue (lock)",
			rb_lock.New[int](rb_lock.WithCapacity[int](c)),
			r,
			w,
		},
		{
			"linked list queue",
			queue.New[int](queue.WithCapacity[int](c)),
			r,
			w,
		},
		{
			"go chan",
			gochan.New[int](gochan.WithCapacity[int](c)),
			r,
			w,
		},
		{
			"linked list stack",
			stack.NewQueueAdapter[int](c),
			r,
			w,
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
	ds      core.Queue[V]
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
					_ = collection.Enqueue(i)
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
