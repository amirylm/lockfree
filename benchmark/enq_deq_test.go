package benchmark

import (
	"testing"
)

func Bench_Int_Concurrent_Single(b *testing.B) {
	BenchEnqueueDequeueInt(b, 128, 1, 1)
}

func Bench_Int_Concurrent_Multi_4(b *testing.B) {
	BenchEnqueueDequeueInt(b, 128, 4, 4)
}

func Bench_Bytes_Concurrent_Single(b *testing.B) {
	BenchEnqueueDequeueBytes(b, 128, 1, 1)
}

func Bench_Concurrent_Multi_4(b *testing.B) {
	BenchEnqueueDequeueBytes(b, 128, 4, 4)
}

func Bench_Concurrent_Multi_16(b *testing.B) {
	BenchEnqueueDequeueBytes(b, 128, 16, 16)
}

func Bench_Concurrent_Multi_Readers(b *testing.B) {
	BenchEnqueueDequeueBytes(b, 128, 8, 2)
}

func Bench_Concurrent_Multi_Writers(b *testing.B) {
	BenchEnqueueDequeueBytes(b, 128, 2, 8)
}
