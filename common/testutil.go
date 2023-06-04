package common

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func DoConcurrencyCheck(pctx context.Context, t *testing.T, coll LockFreeCollection[[]byte], n, readers, writers int) {
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()

	counter := int64(0)
	errDone := errors.New("done")

	writer := Channel(pctx, coll, func(v []byte) error {
		if atomic.AddInt64(&counter, 1) == int64(n) {
			return errDone
		}
		return nil
	}, func(err error) bool {
		if err == errDone {
			cancel()
			return true
		}
		return false
	})

	for i := 0; i < writers; i++ {
		go func() {
			for i := 0; i < n; i++ {
				data := fmt.Sprintf("%x", rand.Intn(100_000_000))
				dataB := []byte(data)
				err := writer(dataB)
				if err == ErrOverflow {
					continue
				}
				require.NoError(t, err)
			}
		}()
	}

	<-ctx.Done()
}
