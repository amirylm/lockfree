package main

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"sync/atomic"

	"github.com/amirylm/lockfree/pool"
)

func main() {
	hashersCreated := atomic.Int32{}

	hashFn := pool.PoolWrapper(func() hash.Hash {
		hashersCreated.Add(1)
		return sha256.New()
	}, func(h hash.Hash, data []byte) [32]byte {
		var b [32]byte
		_, err := h.Write(data)
		defer h.Reset()
		if err != nil {
			return b
		}
		h.Sum(b[:0])
		return b
	})

	n := 10
	done := make(chan struct{})
	for i := 0; i < n; i++ {
		go func(i int) {
			fmt.Printf("%d: %x\n", i+1, hashFn([]byte(fmt.Sprintf("dummy-%d", i))))
			done <- struct{}{}
		}(i)
	}

	for i := 0; i < n; i++ {
		<-done
	}

	fmt.Printf("\nused %d hashers to calculate %d hashed\n", hashersCreated.Load(), n)
}
