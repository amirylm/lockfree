# lockfree

**WARNING: WIP, DO NOT USE in production**

[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://pkg.go.dev/github.com/amirylm/lockfree?tab=doc)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/amirylm/lockfree/blob/main/LICENSE)
![Go version](https://img.shields.io/badge/go-1.20-blue.svg)
![Github Actions](https://github.com/amirylm/lockfree/actions/workflows/lint.yml/badge.svg?branch=main)
![Github Actions](https://github.com/amirylm/lockfree/actions/workflows/test.yml/badge.svg?branch=main)

## Overview

Lock-free data structures implemented with native Golang, based on atomic compare-and-swap operations.
These lock-free data structures are designed to provide concurrent access without relying on traditional locking mechanisms, improving performance and scalability in highly concurrent environments.

### Data Structures

* [x] LL Stack - lock-free stack based on a linked list with `atomic.Pointer` elements.
* [x] LL Queue - lock-free queue based on a linked list with `atomic.Pointer` elements.
* [x] RB Queue - lock-free queue based on a ring buffer that uses a capped slice of `atomic.Pointer` elements.

**NOTE:** lock based data structures were implemented for benchmarking purposes (lock based ring buffer and channel based queue).

### Extras

* [x] Reactor - lock-free reactor that provides a thread-safe, non-blocking, asynchronous event processing. \
It uses lock-free queues for events and control messages.

## Usage

```shell
go get guthub.com/amirylm/lockfree
```

```go
import (
    "github.com/amirylm/lockfree/ringbuffer"
    "github.com/amirylm/lockfree/core"
)

func main() {
    ctx, cancel := context.WithCancel(context.Backgorund())
    defer cancel()

    q := ringbuffer.New[[]byte](256)

    core.Enqueue(ctx, q, []byte("hello ring buffer"))

    val, _ := core.Dequeue(ctx, q)
    fmt.Println(val)
}
```

Detailed examples of usage can be found in `./examples` folder. 
Additionally, you can find more examples in tests and benchmarks.

## Contributing

Contributions to lockfree are welcomed and encouraged! If you find a bug or have an idea for an improvement, please open an issue or submit a pull request.

Before contributing, please make sure to read the [Contributing Guidelines](CONTRIBUTING.md) for more information.

## License

lockfree is open-source software licensed under the [MIT License](LICENSE). Feel free to use, modify, and distribute this library as permitted by the license.
