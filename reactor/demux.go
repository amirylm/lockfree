package reactor

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"sync/atomic"

	"github.com/amirylm/go-options"
	"github.com/amirylm/lockfree/core"
	"github.com/amirylm/lockfree/queue"
	"github.com/amirylm/lockfree/ringbuffer"
)

type Service[T any] interface {
	Select(T) bool
	Handle(T)
}

// Demultiplexer provides a thread-safe, non-blocking, asynchronous event processing.
// Using lock-free queues for events and control messages, and atomic pointers to manage states and workers.
// This component can be used instead of go channels in cases of multiple parallel readers
type Demultiplexer[T any] interface {
	io.Closer
	// Start starts the event loop
	Start(context.Context) error
	// Enqueue adds a new event to the event queue
	Enqueue(T)
	// Register registers handlers. It accepts the event selector, amount of goroutine workers
	// that will be used to process events, and the handlers that will be called.
	Register(id string, s Service[T], workers int)
	// Unregister unregisters handlers
	Unregister(id string)
}

type serviceWrapper[T any] struct {
	id      string
	svc     Service[T]
	workers *atomic.Int32
}

type control int32

const (
	registerService control = iota
	unregisterService
)

type controlEvent[T any] struct {
	control control
	id      string
	svc     Service[T]
	workers int32
}

func WithEventQueue[T any](q core.Queue[T]) options.Option[demultiplexer[T]] {
	return func(r *demultiplexer[T]) {
		r.eventQ = q
	}
}

func WithControlQueue[T any](q core.Queue[controlEvent[T]]) options.Option[demultiplexer[T]] {
	return func(r *demultiplexer[T]) {
		r.controlQ = q
	}
}

func WithCloneFn[T any](f func(T) T) options.Option[demultiplexer[T]] {
	return func(r *demultiplexer[T]) {
		r.cloneFn = f
	}
}

type demultiplexer[T any] struct {
	eventQ   core.Queue[T]
	controlQ core.Queue[controlEvent[T]]

	done    atomic.Pointer[context.CancelFunc]
	cloneFn func(T) T
}

func NewDemux[T any](opts ...options.Option[demultiplexer[T]]) Demultiplexer[T] {
	el := options.Apply(nil, opts...)

	if el.eventQ == nil {
		el.eventQ = ringbuffer.New(ringbuffer.WithCapacity[T](1024))
		eventQType := reflect.TypeOf(el.eventQ)
		fmt.Println("Type of el.eventQ:", eventQType)
	}
	if el.controlQ == nil {
		el.controlQ = queue.New(queue.WithCapacity[controlEvent[T]](32))
		controlQType := reflect.TypeOf(el.controlQ)
		fmt.Println("Type of el.controlQ:", controlQType)
	}
	el.done = atomic.Pointer[context.CancelFunc]{}

	return el
}

func (r *demultiplexer[T]) Close() error {
	cancel := r.done.Load()
	if cancel == nil {
		return nil
	}
	(*cancel)()
	r.done.Store(nil)
	return nil
}

func (r *demultiplexer[T]) Start(pctx context.Context) error {
	ctx, cancel := context.WithCancel(pctx)
	r.done.Store(&cancel)
	services := make([]serviceWrapper[T], 0)
	for ctx.Err() == nil {
		c, ok := r.controlQ.Dequeue()
		if ok {
			services = r.handleControl(services, &c)
			continue
		}
		e, ok := r.eventQ.Dequeue()
		if ok {
			eventServices := r.selectServices(e, services...)
			go r.handleEvent(r.clone(e), eventServices...)
			continue
		}
		runtime.Gosched()
	}

	return ctx.Err()
}

// Register will add the given service (id, selector and handlers).
// Note that we filter existing IDs, one must use Unregister ID before trying to register.
func (r *demultiplexer[T]) Register(serviceID string, service Service[T], workers int) {
	r.controlQ.Enqueue(controlEvent[T]{
		control: registerService,
		id:      serviceID,
		svc:     service,
		workers: int32(workers),
	})
}

func (r *demultiplexer[T]) Unregister(serviceID string) {
	r.controlQ.Enqueue(controlEvent[T]{
		control: unregisterService,
		id:      serviceID,
	})
}

func (r *demultiplexer[T]) Enqueue(t T) {
	r.eventQ.Enqueue(t)
}

func (r *demultiplexer[T]) selectServices(t T, serviceWrappers ...serviceWrapper[T]) []serviceWrapper[T] {
	var selected []serviceWrapper[T]
	for _, s := range serviceWrappers {
		if s.svc.Select(t) {
			selected = append(selected, s)
		}
	}
	return selected
}

// handleEvent handles an event by calling the appropriate handlers.
// Runs in an event thread, and might spawn worker threads.
func (r *demultiplexer[T]) handleEvent(t T, services ...serviceWrapper[T]) {
	for _, s := range services {
		if s.workers.Load() <= 0 {
			// if there are no available workers, run on the event thread
			s.svc.Handle(r.clone(t))
			continue
		}
		s.workers.Add(-1)
		go func(t T, workers *atomic.Int32, svc Service[T]) {
			defer workers.Add(1)
			svc.Handle(r.clone(t))
		}(r.clone(t), s.workers, s.svc)
	}
}

func (r *demultiplexer[T]) handleControl(serviceWrappers []serviceWrapper[T], ce *controlEvent[T]) []serviceWrapper[T] {
	switch ce.control {
	case registerService:
		for _, s := range serviceWrappers {
			if s.id == ce.id {
				return serviceWrappers
			}
		}
		workers := &atomic.Int32{}
		workers.Store(ce.workers)
		return append(serviceWrappers, serviceWrapper[T]{
			id:      ce.id,
			svc:     ce.svc,
			workers: workers,
		})
	case unregisterService:
		updated := make([]serviceWrapper[T], len(serviceWrappers))
		i := 0
		for _, s := range serviceWrappers {
			if s.id != ce.id {
				updated[i] = s
				i++
			}
		}
		if i == 0 {
			return []serviceWrapper[T]{}
		}
		return updated[:i]
	default:
		return serviceWrappers
	}
}

func (r *demultiplexer[T]) clone(t T) T {
	if r.cloneFn != nil {
		return r.cloneFn(t)
	}
	return t
}
