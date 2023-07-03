package eventloop

import (
	"context"
	"runtime"
	"sync/atomic"

	"github.com/amirylm/lockfree/core"
	"github.com/amirylm/lockfree/ringbuffer"
)

// Handler is a function that handles events
type Handler[T any] func(T)

// Selector is a predicate that selects events for a given set of handlers
type Selector[T any] func(T) bool

// EventLoop is a lock-free event loop that provides
// a thread-safe, non-blocking, asynchronous event processing.
// It is based on the reactor pattern and
// uses lock-free queues for events and control messages.
type EventLoop[T any] interface {
	// Start starts the event loop
	Start(context.Context) error
	// Enqueue adds a new event to the event loop
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

type eventLoop[T any] struct {
	eventQ   core.Queue[T]
	controlQ core.Queue[controlEvent[T]]
	services []service[T]

	workers atomic.Int32
}

func New[T any](workers int, q core.Queue[T]) *eventLoop[T] {
	if q == nil {
		q = ringbuffer.NewWithOverride[T](256)
	}
	el := &eventLoop[T]{
		eventQ:   q,
		controlQ: ringbuffer.New[controlEvent[T]](32),
		services: []service[T]{},
		workers:  atomic.Int32{},
	}

	el.workers.Store(int32(workers))

	return el
}

func (el *eventLoop[T]) Start(pctx context.Context) error {
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()

	var services []service[T]
	for ctx.Err() == nil {
		c, ok := el.controlQ.Dequeue()
		if ok {
			el.handleControl(c)
			services = make([]service[T], len(el.services))
			copy(services, el.services)
			continue
		}
		e, ok := el.eventQ.Dequeue()
		if ok {
			go el.handleEvent(e, services...)
			continue
		}
		runtime.Gosched()
	}

	return ctx.Err()
}

func (el *eventLoop[T]) Register(serviceID string, selector Selector[T], workers int, handlers ...Handler[T]) {
	el.controlQ.Enqueue(controlEvent[T]{
		control: registerService,
		svc: service[T]{
			id:       serviceID,
			handlers: handlers,
			selector: selector,
			workers:  workers,
		},
	})
}

func (el *eventLoop[T]) Unregister(serviceID string) {
	el.controlQ.Enqueue(controlEvent[T]{
		control: unregisterService,
		svc: service[T]{
			id: serviceID,
		},
	})
}

func (el *eventLoop[T]) Enqueue(t T) {
	el.eventQ.Enqueue(t)
}

// handleEvent handles an event by calling the appropriate handlers.
// Runs in an event thread, and might spawn worker threads.
func (el *eventLoop[T]) handleEvent(t T, services ...service[T]) {
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

func (el *eventLoop[T]) handleControl(ce controlEvent[T]) {
	switch ce.control {
	case registerService:
		for _, svc := range el.services {
			if svc.id == ce.svc.id {
				return
			}
		}
		el.services = append(el.services, ce.svc)
	case unregisterService:
		cleaned := make([]service[T], len(el.services))
		i := 0
		for _, svc := range el.services {
			if svc.id != ce.svc.id {
				cleaned[i] = svc
				i++
			}
		}
		el.services = cleaned[:i]
	}
}
