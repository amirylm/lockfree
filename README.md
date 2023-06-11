# lockfree

**WARNING: WIP**

[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://pkg.go.dev/github.com/amirylm/lockfree?tab=doc)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/amirylm/lockfree/blob/main/LICENSE)
![Go version](https://img.shields.io/badge/go-1.20-blue.svg)
![Github Actions](https://github.com/amirylm/lockfree/actions/workflows/lint.yml/badge.svg?branch=main)
![Github Actions](https://github.com/amirylm/lockfree/actions/workflows/test.yml/badge.svg?branch=main)

## Overview

Lock-free data structures implemented with native Golang, based on atomic compare-and-swap operations.
These lock-free data structures are designed to provide concurrent access without the need for traditional locking mechanisms, improving performance and scalability in highly concurrent environments.

### Data Structures

* [x] LL Stack - lock-free stack based on a linked list with `atomic.Pointer` elements.
* [x] LL Queue - lock-free queue based on a linked list with `atomic.Pointer` elements.
* [x] RB Queue - lock-free queue based on a ring buffer that uses a capped slice of `atomic.Pointer` elements.

**NOTE:** corresponding lock based data structures were implemented for benchmarking (lock based ring buffer and channel based queue).

### Benchmarks

```
go test -benchmem -bench . ./benchmark/...

goos: darwin
goarch: amd64
pkg: github.com/amirylm/lockfree/benchmark
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
```

Single threaded:
```
Benchmark_PushPopInt/ring_buffer-8           25576218   46.38 ns/op   8 B/op    1 allocs/op
Benchmark_PushPopInt/queue-8                 46466366   21.91 ns/op   8 B/op    1 allocs/op
Benchmark_PushPopInt/ring_buffer_(lock)-8    15942847   65.96 ns/op   0 B/op    0 allocs/op
Benchmark_PushPopInt/go_channel-8            21413254   48.90 ns/op   0 B/op    0 allocs/op
```

Multiple goroutines, with a large set of readers:
```
Benchmark_PushPopBytes_Concurrent_Multi_Readers_64/ring_buffer_(atomic)_64_R_4_W-8   60691   20163 ns/op   3808 B/op   140 allocs/op
Benchmark_PushPopBytes_Concurrent_Multi_Readers_64/queue_64_R_4_W-8                  61374   20269 ns/op   3808 B/op   140 allocs/op
Benchmark_PushPopBytes_Concurrent_Multi_Readers_64/ring_buffer_(lock)_64_R_4_W-8     60181   20041 ns/op   3776 B/op   136 allocs/op
Benchmark_PushPopBytes_Concurrent_Multi_Readers_64/go_channel_64_R_4_W-8             10000   165496 ns/op  38210 B/op  255 allocs/op
```

With a large set of writers:
```
Benchmark_PushPopBytes_Concurrent_Multi_Writers_64/ring_buffer_(atomic)_4_R_64_W-8  60235   19996 ns/op   3809 B/op   200 allocs/op
Benchmark_PushPopBytes_Concurrent_Multi_Writers_64/queue_4_R_64_W-8                 60631   21435 ns/op   4831 B/op   263 allocs/op
Benchmark_PushPopBytes_Concurrent_Multi_Writers_64/ring_buffer_(lock)_4_R_64_W-8    59246   19781 ns/op   3296 B/op   136 allocs/op
Benchmark_PushPopBytes_Concurrent_Multi_Writers_64/go_channel_4_R_64_W-8            57892   22307 ns/op   3296 B/op   136 allocs/op
```

## Usage

You can find examples of usage in `./examples` folder, tests, and benchmarks.

## Contributing

Contributions to lockfree are welcome and encouraged! If you find a bug or have an idea for an improvement, please open an issue or submit a pull request.

Before contributing, please make sure to read the [Contributing Guidelines](CONTRIBUTING.md) for more information.

## License

lockfree is open-source software licensed under the [MIT License](LICENSE). Feel free to use, modify, and distribute this library as permitted by the license.