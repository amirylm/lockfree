package common

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

type Collection[Value any] interface {
	Store(Value) error
	Read() (*Value, error)
	Len() int
	IsEmpty() bool
}

func DoConcurrencyCheck(pctx context.Context, t *testing.T, coll Collection[[]byte], n, readers, writers int) {
	counter := int64(0)

	for i := 0; i < writers; i++ {
		go func() {
			for i := 0; i < n; i++ {
				data := fmt.Sprintf("%x", rand.Intn(100_000_000))
				dataB := []byte(data)
				require.NoError(t, coll.Store(dataB))
			}
		}()
	}

	var wg sync.WaitGroup
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pctx.Err() == nil {
				if coll.IsEmpty() {
					runtime.Gosched()
					continue
				}
				_, err := coll.Read()
				require.NoError(t, err)
				if atomic.AddInt64(&counter, 1) == int64(n) {
					return
				}
			}
		}()
	}

	wg.Wait()

	require.Equal(t, int64(n), atomic.LoadInt64(&counter))
}
