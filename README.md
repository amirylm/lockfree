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

### Features

The lockfree library currently includes the following data structures:

#### 1. LL Stack

A lock-free stack implementation that is based on a linked list with `atomic.Pointer` elements.

#### 2. LL Queue

A lock-free queue implementation that is based on a linked list with `atomic.Pointer` elements.

#### 3. RB Queue 

A lock-free queue implementation based on a ring buffer that uses an underlying capped slice of `atomic.Pointer` elements.

### Benchmarks

**TODO**

## Usage

You can find examples of usage in `./examples` folder, tests, and benchmarks.

## Contributing

Contributions to lockfree are welcome and encouraged! If you find a bug or have an idea for an improvement, please open an issue or submit a pull request.

Before contributing, please make sure to read the [Contributing Guidelines](CONTRIBUTING.md) for more information.

## License

lockfree is open-source software licensed under the [MIT License](LICENSE). Feel free to use, modify, and distribute this library as permitted by the license.