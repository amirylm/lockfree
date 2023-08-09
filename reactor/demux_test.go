package reactor

import (
	"context"
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amirylm/lockfree/queue"
	"github.com/stretchr/testify/require"
)

func TestDemux(t *testing.T) {
	d := NewDemux(WithCloneFn(func(b []byte) []byte {
		if len(b) == 0 {
			return b
		}
		cp := make([]byte, len(b))
		copy(b, cp)
		return cp
	})).(*demultiplexer[[]byte])
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	go func() {
		t.Log("starting reactor")
		_ = d.Start(ctx)
	}()

	t.Run("sanity", func(t *testing.T) {
		expectedCount := int32(3)
		cs := NewCountService()
		d.Register("test", cs, 0)
		defer d.Unregister("test")

		// double register should not work
		d.Register("test", cs, 0)

		go d.Enqueue([]byte("hello"))
		go d.Enqueue([]byte("world"))
		go d.Enqueue([]byte("hello-world"))

		for cs.getCount() < expectedCount && ctx.Err() == nil {
			runtime.Gosched()
		}
		require.Equal(t, expectedCount, cs.getCount())
	})

	t.Run("multi-services", func(t *testing.T) {
		cs := NewCountService()
		expectedCount := int32(3 * 3)
		d.Register("test-workers", cs, 2)
		d.Register("test-workers2", cs, 2)
		d.Register("test-workers3", cs, 2)
		defer d.Unregister("test-workers")
		defer d.Unregister("test-workers2")
		defer d.Unregister("test-workers3")

		cs2 := NewCountService()
		expectedCount2 := int32(3)
		d.Register("test-workers-4", cs2, 2)
		defer d.Unregister("test-workers-4")

		d.Enqueue([]byte("hello"))
		d.Enqueue([]byte("world"))
		d.Enqueue([]byte("hello-world"))
		for cs.getCount() < expectedCount && ctx.Err() == nil {
			runtime.Gosched()
		}
		require.Equal(t, expectedCount, cs.getCount())
		require.Equal(t, expectedCount2, cs2.getCount())
	})

	t.Run("close", func(t *testing.T) {
		require.NoError(t, d.Close())
		require.NoError(t, d.Close())
		d.Enqueue([]byte("hello"))
	})
}

type TestService struct{}

func (bs *TestService) Select(data []byte) bool {
	return true
}

func (bs *TestService) Handle(data []byte) {
	fmt.Println("Handling data:", data)
}

func TestDemux_handleControl(t *testing.T) {
	r := NewDemux(
		WithEventQueue(queue.New(queue.WithCapacity[[]byte](32))),
		WithControlQueue(queue.New(queue.WithCapacity[controlEvent[[]byte]](32))),
	).(*demultiplexer[[]byte])
	test_service := &TestService{}
	var atomic_workers atomic.Int32
	var atomic_workers_2 atomic.Int32
	atomic_workers.Store(0)
	atomic_workers_2.Store(2)
	tests := []struct {
		name     string
		existing []serviceWrapper[[]byte]
		events   []controlEvent[[]byte]
		want     []serviceWrapper[[]byte]
	}{
		{
			name:     "empty",
			existing: nil,
			events: []controlEvent[[]byte]{
				{
					control: registerService,
					id:      "test",
					workers: 0,
				},
			},
			want: []serviceWrapper[[]byte]{
				{
					id:      "test",
					workers: &atomic_workers,
				},
			},
		},
		{
			name: "unknown control",
			existing: []serviceWrapper[[]byte]{
				{
					id: "test",
				},
			},
			events: []controlEvent[[]byte]{
				{
					control: 5,
					svc:     test_service,
				},
			},
			want: []serviceWrapper[[]byte]{
				{
					id: "test",
				},
			},
		},
		{
			name: "register",
			existing: []serviceWrapper[[]byte]{
				{
					id:      "test",
					workers: &atomic_workers,
				},
			},
			events: []controlEvent[[]byte]{
				{
					id:      "test2",
					control: registerService,
					svc:     test_service,
					workers: 0,
				},
				{
					id:      "test3",
					control: registerService,
					svc:     test_service,
					workers: 2,
				},
			},
			want: []serviceWrapper[[]byte]{
				{
					id:      "test",
					workers: &atomic_workers,
				},
				{
					id:      "test2",
					svc:     test_service,
					workers: &atomic_workers,
				},
				{
					id:      "test3",
					svc:     test_service,
					workers: &atomic_workers_2,
				},
			},
		},
		{
			name: "double register",
			existing: []serviceWrapper[[]byte]{
				{
					id: "test",
				},
			},
			events: []controlEvent[[]byte]{
				{
					id:      "test",
					control: registerService,
					svc:     test_service,
				},
			},
			want: []serviceWrapper[[]byte]{
				{
					id: "test",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			services := tc.existing
			for _, e := range tc.events {
				services = r.handleControl(services, &e)
			}
			require.Equal(t, tc.want, services)
		})
	}
}

func (cs CountService) Count() int32 {
	cs.Counter.Add(1)
	return cs.Counter.Load()
}

type CountService struct {
	Counter *atomic.Int32
}

func NewCountService() *CountService {
	return &CountService{
		Counter: &atomic.Int32{},
	}
}

func (cs CountService) Select(b []byte) bool {
	cs.Count()
	return true
}

func (cs CountService) getCount() int32 {
	return cs.Counter.Load()
}
func (cs CountService) Handle(b []byte) {
}
