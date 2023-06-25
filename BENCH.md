# Benchmarking - lockfree

## Usage

```shell
make bench

> go test -benchmem -bench . ./benchmark/...
```

## Analysis

The benchmarks were taken on the following machine:
```
goos: darwin
goarch: amd64
pkg: github.com/amirylm/lockfree/benchmark
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
```

#### Single threaded

```
Benchmark_PushPopInt/ring_buffer-8           25576218   46.38 ns/op   8 B/op    1 allocs/op
Benchmark_PushPopInt/queue-8                 46466366   21.91 ns/op   8 B/op    1 allocs/op
Benchmark_PushPopInt/ring_buffer_(lock)-8    15942847   65.96 ns/op   0 B/op    0 allocs/op
Benchmark_PushPopInt/go_channel-8            21413254   48.90 ns/op   0 B/op    0 allocs/op
```

#### Multiple goroutines, more readers

```
Benchmark_PushPopBytes_Concurrent_Multi_Readers_64/ring_buffer_(atomic)_64_R_4_W-8   60691   20163 ns/op   3808 B/op   140 allocs/op
Benchmark_PushPopBytes_Concurrent_Multi_Readers_64/queue_64_R_4_W-8                  61374   20269 ns/op   3808 B/op   140 allocs/op
Benchmark_PushPopBytes_Concurrent_Multi_Readers_64/ring_buffer_(lock)_64_R_4_W-8     60181   20041 ns/op   3776 B/op   136 allocs/op
Benchmark_PushPopBytes_Concurrent_Multi_Readers_64/go_channel_64_R_4_W-8             10000   165496 ns/op  38210 B/op  255 allocs/op
```

#### Multiple goroutines, more writers

```
Benchmark_PushPopBytes_Concurrent_Multi_Writers_64/ring_buffer_(atomic)_4_R_64_W-8  60235   19996 ns/op   3809 B/op   200 allocs/op
Benchmark_PushPopBytes_Concurrent_Multi_Writers_64/queue_4_R_64_W-8                 60631   21435 ns/op   4831 B/op   263 allocs/op
Benchmark_PushPopBytes_Concurrent_Multi_Writers_64/ring_buffer_(lock)_4_R_64_W-8    59246   19781 ns/op   3296 B/op   136 allocs/op
Benchmark_PushPopBytes_Concurrent_Multi_Writers_64/go_channel_4_R_64_W-8            57892   22307 ns/op   3296 B/op   136 allocs/op
```

