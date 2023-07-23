package reactor

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/amirylm/go-options"
)

type Reactor[T any] interface {
	io.Closer
	Start(pctx context.Context) error

	Enqueue(events ...T)
	EnqueueWait(context.Context, T) (T, error)

	AddHandler(string, Selector[T], int, EventHandler[T])
	RemoveHandler(string)

	AddCallback(string, Selector[T], int, DemuxHandler[T])
	RemoveCallback(string)
}

// EventHandler is a function that handles events, it accepts a callback function as a second parameter.
// The callback function is expected to be called once the event was processed
type EventHandler[T any] func(T, func(T, error))

func WithEventsDemux[T any](d Demultiplexer[Event[T]]) options.Option[reactor[T]] {
	return func(r *reactor[T]) {
		r.events = d
	}
}

func WithCallbacksDemux[T any](d Demultiplexer[Event[T]]) options.Option[reactor[T]] {
	return func(r *reactor[T]) {
		r.callbacks = d
	}
}

func WithTimes[T any](tick, timeout time.Duration) options.Option[reactor[T]] {
	return func(r *reactor[T]) {
		r.tick = tick
		r.timeout = timeout
	}
}

func New[T any](opts ...options.Option[reactor[T]]) Reactor[T] {
	r := options.Apply(nil, opts...)

	if r.events == nil {
		r.events = NewDemux[Event[T]]()
	}
	if r.callbacks == nil {
		r.callbacks = NewDemux[Event[T]]()
	}
	if r.tick == 0 {
		r.tick = time.Second / 2
	}
	if r.timeout == 0 {
		r.timeout = time.Second * 10
	}

	return r
}

type reactor[T any] struct {
	events, callbacks Demultiplexer[Event[T]]
	tick, timeout     time.Duration

	done atomic.Pointer[context.CancelFunc]
}

func (r *reactor[T]) genID(T) ID {
	return []byte(fmt.Sprintf("%04d-%08d-%04d",
		rand.Intn(9999), rand.Intn(99999999), rand.Intn(9999)))
}

func (r *reactor[T]) Start(pctx context.Context) error {
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()
	r.done.Store(&cancel)
	go func() {
		_ = r.callbacks.Start(ctx)
	}()
	return r.events.Start(ctx)
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

func (r *reactor[T]) Enqueue(events ...T) {
	for _, data := range events {
		r.events.Enqueue(Event[T]{
			ID:    r.genID(data),
			Nonce: 0,
			Data:  data,
		})
	}
}

func (r *reactor[T]) EnqueueWait(pctx context.Context, data T) (T, error) {
	ctx, cancel := context.WithTimeout(pctx, r.timeout)
	defer cancel()

	resultp := atomic.Pointer[Event[T]]{}
	nonce := int64(1)
	id := r.genID(data)

	cid := fmt.Sprintf("%x:%d", id, nonce)
	r.callbacks.Register(cid, func(e Event[T]) bool {
		return bytes.Equal(e.ID, id) && e.Nonce == nonce
	}, 0, func(e Event[T]) {
		resultp.Store(&e)
	})
	defer r.callbacks.Unregister(cid)

	r.events.Enqueue(Event[T]{
		ID:    id,
		Nonce: nonce - 1,
		Data:  data,
	})

	result := resultp.Load()
	for ctx.Err() == nil && result == nil {
		time.Sleep(r.tick)
		result = resultp.Load()
	}

	if result != nil {
		return result.Data, result.Err
	}
	var res T
	return res, ctx.Err()
}

func (r *reactor[T]) AddHandler(id string, selector Selector[T], workers int, handler EventHandler[T]) {
	r.events.Register(id, func(e Event[T]) bool {
		return selector(e.Data)
	}, workers, func(e Event[T]) {
		n := e.Nonce
		eid := e.ID
		callbacks := r.callbacks
		handler(e.Data, func(data T, err error) {
			resp := Event[T]{
				ID:    eid,
				Nonce: n + 1,
				Data:  data,
			}
			if err != nil {
				resp.Err = err
			}
			callbacks.Enqueue(resp)
		})
	})
}

func (r *reactor[T]) RemoveHandler(id string) {
	r.events.Unregister(id)
}

func (r *reactor[T]) AddCallback(id string, selector Selector[T], workers int, handler DemuxHandler[T]) {
	r.callbacks.Register(id, func(e Event[T]) bool {
		return selector(e.Data)
	}, workers, func(e Event[T]) {
		handler(e.Data)
	})
}

func (r *reactor[T]) RemoveCallback(id string) {
	r.callbacks.Unregister(id)
}

// ID is the ID used for events
type ID []byte

func (id ID) String() string {
	return hex.EncodeToString(id)
}

func IDFromString(idstr string) ID {
	id, err := hex.DecodeString(idstr)
	if err != nil {
		return nil
	}
	return id
}

type Event[T any] struct {
	ID    ID
	nonce int64
	Data  T
	Err   error
}

func (e Event[T]) Nonce() int64 {
	return e.nonce
}
