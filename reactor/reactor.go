package reactor

import (
	"context"
	"runtime"
	"sync/atomic"

	"github.com/amirylm/go-options"
	"github.com/amirylm/lockfree/core"
	"github.com/amirylm/lockfree/ringbuffer"
)

// Handler is a function that handles events
type Handler[T any] func(T)

// Selector is a predicate that selects events for a given set of handlers
type Selector[T any] func(T) bool

// Reactor provides a thread-safe, non-blocking, asynchronous event processing.
// It uses lock-free queues for events and control messages.
type Reactor[T any] interface {
	// Start starts the event loop
	Start(context.Context) error
	// Enqueue adds a new event to the event queue
	Enqueue(T)
	// Register registers handlers. It accepts the event selector, amount of goroutine workers
	// that will be used to process events, and the handlers that will be called.
	Register(id string, s Selector[T], workers int, handlers ...Handler[T])
	// Unregister unregisters handlers
	Unregister(id string)
}

type service[T any] struct {
	id       string
	handlers []Handler[T]
	selector Selector[T]
	workers  int
}

type control int32

const (
	registerService control = iota
	unregisterService
)

type controlEvent[T any] struct {
	control control
	svc     service[T]
}

func WithEventQueue[T any](q core.Queue[T]) options.Option[reactor[T]] {
	return func(r *reactor[T]) {
		r.eventQ = q
	}
}

func WithControlQueue[T any](q core.Queue[controlEvent[T]]) options.Option[reactor[T]] {
	return func(r *reactor[T]) {
		r.controlQ = q
	}
}

type reactor[T any] struct {
	eventQ   core.Queue[T]
	controlQ core.Queue[controlEvent[T]]

	done atomic.Pointer[context.CancelFunc]
}

func New[T any](opts ...options.Option[reactor[T]]) *reactor[T] {
	el := options.Apply(nil, opts...)

	if el.eventQ == nil {
		el.eventQ = ringbuffer.New(ringbuffer.WithCapacity[T](256), ringbuffer.WithOverride[T]())
	}
	if el.controlQ == nil {
		el.controlQ = ringbuffer.New(ringbuffer.WithCapacity[controlEvent[T]](32))
	}
	el.done = atomic.Pointer[context.CancelFunc]{}

	return el
}

func (r *reactor[T]) Close() error {
	cancel := r.done.Load()
	if cancel == nil {
		return nil
	}
	(*cancel)()
	r.done.Store(nil)
	return nil
}

func (r *reactor[T]) Start(pctx context.Context) error {
	ctx, cancel := context.WithCancel(pctx)
	r.done.Store(&cancel)

	var services []service[T]
	for ctx.Err() == nil {
		c, ok := r.controlQ.Dequeue()
		if ok {
			services = r.handleControl(services, c)
			continue
		}
		e, ok := r.eventQ.Dequeue()
		if ok {
			go r.handleEvent(e, services...)
			continue
		}
		runtime.Gosched()
	}

	return ctx.Err()
}

func (r *reactor[T]) Register(serviceID string, selector Selector[T], workers int, handlers ...Handler[T]) {
	r.controlQ.Enqueue(controlEvent[T]{
		control: registerService,
		svc: service[T]{
			id:       serviceID,
			handlers: handlers,
			selector: selector,
			workers:  workers,
		},
	})
}

func (r *reactor[T]) Unregister(serviceID string) {
	r.controlQ.Enqueue(controlEvent[T]{
		control: unregisterService,
		svc: service[T]{
			id: serviceID,
		},
	})
}

func (r *reactor[T]) Enqueue(t T) {
	r.eventQ.Enqueue(t)
}

// handleEvent handles an event by calling the appropriate handlers.
// Runs in an event thread, and might spawn worker threads.
func (r *reactor[T]) handleEvent(t T, services ...service[T]) {
	for _, svc := range services {
		if svc.selector(t) {
			workers := atomic.Int32{}
			workers.Store(int32(svc.workers))
			for _, h := range svc.handlers {
				if workers.Load() == 0 {
					// if there are no available workers, run on the event thread
					h(t)
					continue
				}
				workers.Add(-1)
				// run on a worker thread
				go func(h Handler[T], t T) {
					defer workers.Add(1)
					h(t)
				}(h, t)
			}
		}
	}
}

func (r *reactor[T]) handleControl(services []service[T], ce controlEvent[T]) []service[T] {
	switch ce.control {
	case registerService:
		for _, svc := range services {
			if svc.id == ce.svc.id {
				return services
			}
		}
		return append(services, ce.svc)
	case unregisterService:
		updated := make([]service[T], len(services))
		i := 0
		for _, svc := range services {
			if svc.id != ce.svc.id {
				updated[i] = svc
				i++
			}
		}
		return updated[:i]
	default:
		return services
	}
}
