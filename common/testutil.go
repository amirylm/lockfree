package common

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

type ElementGenerator[Value any] func(i int) Value

type ElementAssertor[Value any] func(i int, v Value) bool

func SanityTest[Value any](t *testing.T, n int, factory DataStructureFactory[Value], gen ElementGenerator[Value], assertor ElementAssertor[Value]) {
	ds := factory(n)
	require.True(t, ds.Empty(), "should be empty")
	for i := 0; i < n; i++ {
		require.True(t, ds.Push(gen(i)), "failed to push element in index %d", i)
	}
	require.Equal(t, n, ds.Size(), "didn't push all elements")
	require.True(t, ds.Full(), "should be full")
	require.False(t, ds.Empty(), "should'nt be empty")
	require.False(t, ds.Push(gen(n)), "shouldn't be able to push when full")
	i := 0
	for !ds.Empty() {
		val, ok := ds.Pop()
		require.True(t, ok, "failed to pop")
		require.True(t, assertor(i, val), "assertion failed: element %d with value %+v", i, val)
		i++
	}
	require.Equal(t, n, i)
	require.Equal(t, 0, ds.Size(), "didn't removed all elements")
	require.False(t, ds.Full(), "shouldn't be full")
	require.True(t, ds.Empty(), "should be empty")
	require.True(t, ds.Push(gen(n)), "should be able to push elements after pop")
}

func ConcurrencyTest[Value any](t *testing.T, pctx context.Context, cap, n, readers, writers int, factory DataStructureFactory[Value], gen ElementGenerator[Value], assertor ElementAssertor[Value]) (int64, int64) {
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()

	var reads, writes int64

	ds := factory(cap)

	var wg sync.WaitGroup

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < n || ctx.Err() != nil; i++ {
				element := gen(i)
				for !ds.Push(element) {
					if ctx.Err() != nil {
						return
					}
					runtime.Gosched()
				}
				atomic.AddInt64(&writes, 1)
			}
		}()
	}

	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < n || ctx.Err() != nil; i++ {
				element, ok := ds.Pop()
				for !ok {
					if ctx.Err() != nil {
						return
					}
					runtime.Gosched()
					element, ok = ds.Pop()
				}
				require.True(t, assertor(i, element), "assertion failed: element %d with value %+v", i, element)
				atomic.AddInt64(&reads, 1)
			}
		}()
	}

	wg.Wait()

	return reads, writes
}
