package ringbuffer

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRingBuffer_EnqueueDequeue(t *testing.T) {
	nitems := 10
	rb := New[int](nitems)
	for i := 0; i < nitems; i++ {
		d := i + 1
		require.NoError(t, rb.Enqueue(d))
	}
	require.Error(t, rb.Enqueue(11))
	require.Equal(t, nitems, rb.Len())
	for i := 0; i < nitems; i++ {
		val, ok := rb.Dequeue()
		require.NoError(t, ok, i)
		require.NotNil(t, val, i)
		require.Equal(t, i+1, val)
		require.Equal(t, nitems-(i+1), rb.Len())
	}
	require.Equal(t, 0, rb.Len())
	require.NoError(t, rb.Enqueue(1))
}

func TestRingBuffer_Concurrency(t *testing.T) {
	pctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*2))
	defer cancel()

	sendSig := make(chan bool, 4)

	n := 100
	q := New[[]byte](128)
	counter := int64(0)

	go func() {
		ctx, cancel := context.WithCancel(pctx)
		defer cancel()

		for {
			select {
			case <-sendSig:
				data := fmt.Sprintf("%x", rand.Intn(100_000_000))
				dataB := []byte(data)
				require.NoError(t, q.Enqueue(dataB))
			case <-ctx.Done():
				return
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for pctx.Err() == nil {
			if q.IsEmpty() {
				runtime.Gosched()
				continue
			}
			_, err := q.Dequeue()
			require.NoError(t, err)
			if atomic.AddInt64(&counter, 1) == int64(n) {
				return
			}
		}
	}()

	for i := 0; i < n; i++ {
		sendSig <- true
	}

	wg.Wait()

	require.Equal(t, int64(n), atomic.LoadInt64(&counter))
}
