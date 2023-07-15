package reactor

import (
	"context"
	"io"
	"runtime"
	"sync/atomic"

	"github.com/amirylm/go-options"
	"github.com/amirylm/lockfree/core"
	"github.com/amirylm/lockfree/ringbuffer"
)

// DemuxHandler is a function that handles events
type DemuxHandler[T any] func(T)

// Selector is a predicate that selects events for a given set of handlers
type Selector[T any] func(T) bool

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
	Register(id string, s Selector[T], workers int, handlers ...DemuxHandler[T])
	// Unregister unregisters handlers
	Unregister(id string)
}

type service[T any] struct {
	id       string
	handlers []DemuxHandler[T]
	selector Selector[T]
	workers  *atomic.Int32
}

type control int32

const (
	registerService control = iota
	unregisterService
)

type controlEvent[T any] struct {
	control control

	svc service[T]
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
		el.eventQ = ringbuffer.New(ringbuffer.WithCapacity[T](1024), ringbuffer.WithOverride[T]())
	}
	if el.controlQ == nil {
		el.controlQ = ringbuffer.New(ringbuffer.WithCapacity[controlEvent[T]](32))
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

	var services []service[T]
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
func (r *demultiplexer[T]) Register(serviceID string, selector Selector[T], workers int, handlers ...DemuxHandler[T]) {
	if len(handlers) == 0 {
		return
	}
	aworkers := &atomic.Int32{}
	aworkers.Store(int32(workers))
	r.controlQ.Enqueue(controlEvent[T]{
		control: registerService,
		svc: service[T]{
			id:       serviceID,
			handlers: handlers,
			selector: selector,
			workers:  aworkers,
		},
	})
}

func (r *demultiplexer[T]) Unregister(serviceID string) {
	r.controlQ.Enqueue(controlEvent[T]{
		control: unregisterService,
		svc: service[T]{
			id: serviceID,
		},
	})
}

func (r *demultiplexer[T]) Enqueue(t T) {
	r.eventQ.Enqueue(t)
}

func (r *demultiplexer[T]) selectServices(t T, services ...service[T]) []service[T] {
	var selected []service[T]
	for _, svc := range services {
		if svc.selector(t) {
			selected = append(selected, svc)
		}
	}
	return selected
}

// handleEvent handles an event by calling the appropriate handlers.
// Runs in an event thread, and might spawn worker threads.
func (r *demultiplexer[T]) handleEvent(t T, services ...service[T]) {
	for _, svc := range services {
		if svc.workers.Load() <= 0 {
			// if there are no available workers, run on the event thread
			r.invokeHandlers(r.clone(t), svc.handlers...)
			continue
		}
		svc.workers.Add(-1)
		go func(t T, workers *atomic.Int32, handlers ...DemuxHandler[T]) {
			defer workers.Add(1)
			r.invokeHandlers(t, handlers...)
		}(r.clone(t), svc.workers, svc.handlers...)
	}
}

func (r *demultiplexer[T]) handleControl(services []service[T], ce *controlEvent[T]) []service[T] {
	switch ce.control {
	case registerService:
		for _, svc := range services {
			if svc.id == ce.svc.id {
				return services
			}
		}
		return append(services, service[T]{
			id:       ce.svc.id,
			handlers: ce.svc.handlers,
			selector: ce.svc.selector,
			workers:  ce.svc.workers,
		})
	case unregisterService:
		updated := make([]service[T], len(services))
		i := 0
		for _, svc := range services {
			if svc.id != ce.svc.id {
				updated[i] = svc
				i++
			}
		}
		if i == 0 {
			return []service[T]{}
		}
		return updated[:i]
	default:
		return services
	}
}

func (r *demultiplexer[T]) invokeHandlers(t T, handlers ...DemuxHandler[T]) {
	for _, h := range handlers {
		h(t)
	}
}

func (r *demultiplexer[T]) clone(t T) T {
	if r.cloneFn != nil {
		return r.cloneFn(t)
	}
	return t
}
