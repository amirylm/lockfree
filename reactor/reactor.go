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

type Reactor[E, C any] interface {
	io.Closer
	Start(pctx context.Context) error

	Enqueue(events ...E)
	EnqueueWait(context.Context, E) (C, error)

	AddHandler(string, Selector[E], int, EventHandler[E, C])
	RemoveHandler(string)

	AddCallback(string, Selector[C], int, DemuxHandler[C])
	RemoveCallback(string)
}

// EventHandler is a function that handles events, it accepts a callback function as a second parameter.
// The callback function is expected to be called once the event was processed
type EventHandler[T, C any] func(T, func(C, error))

func WithEventsDemux[T, C any](d Demultiplexer[Event[T]]) options.Option[reactor[T, C]] {
	return func(r *reactor[T, C]) {
		r.events = d
	}
}

func WithCallbacksDemux[T, C any](d Demultiplexer[Event[C]]) options.Option[reactor[T, C]] {
	return func(r *reactor[T, C]) {
		r.callbacks = d
	}
}

func WithTimes[T, C any](tick, timeout time.Duration) options.Option[reactor[T, C]] {
	return func(r *reactor[T, C]) {
		r.tick = tick
		r.timeout = timeout
	}
}

func New[T, C any](opts ...options.Option[reactor[T, C]]) Reactor[T, C] {
	r := options.Apply(nil, opts...)

	if r.events == nil {
		r.events = NewDemux[Event[T]]()
	}
	if r.callbacks == nil {
		r.callbacks = NewDemux[Event[C]]()
	}
	if r.tick == 0 {
		r.tick = time.Second / 2
	}
	if r.timeout == 0 {
		r.timeout = time.Second * 10
	}

	return r
}

type reactor[T, C any] struct {
	events        Demultiplexer[Event[T]]
	callbacks     Demultiplexer[Event[C]]
	tick, timeout time.Duration

	done atomic.Pointer[context.CancelFunc]
}

func (r *reactor[T, C]) genID(T) ID {
	return []byte(fmt.Sprintf("%04d-%08d-%04d",
		rand.Intn(9999), rand.Intn(99999999), rand.Intn(9999)))
}

func (r *reactor[T, C]) Start(pctx context.Context) error {
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()
	r.done.Store(&cancel)
	go func() {
		_ = r.callbacks.Start(ctx)
	}()
	return r.events.Start(ctx)
}

func (r *reactor[T, C]) Close() error {
	cancel := r.done.Load()
	if cancel == nil {
		return nil
	}
	(*cancel)()
	r.done.Store(nil)
	return nil
}

func (r *reactor[T, C]) Enqueue(events ...T) {
	for _, data := range events {
		r.events.Enqueue(Event[T]{
			ID:    r.genID(data),
			nonce: 0,
			Data:  data,
		})
	}
}

func (r *reactor[T, C]) EnqueueWait(pctx context.Context, data T) (C, error) {
	ctx, cancel := context.WithTimeout(pctx, r.timeout)
	defer cancel()

	resultp := atomic.Pointer[Event[C]]{}
	nonce := int64(1)
	id := r.genID(data)

	cid := fmt.Sprintf("%x:%d", id, nonce)
	r.callbacks.Register(cid, func(e Event[C]) bool {
		return bytes.Equal(e.ID, id) && e.nonce == nonce
	}, 0, func(e Event[C]) {
		resultp.Store(&e)
	})
	defer r.callbacks.Unregister(cid)

	r.events.Enqueue(Event[T]{
		ID:    id,
		nonce: nonce - 1,
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
	var res C
	return res, ctx.Err()
}

func (r *reactor[T, C]) AddHandler(id string, selector Selector[T], workers int, handler EventHandler[T, C]) {
	r.events.Register(id, func(e Event[T]) bool {
		return selector(e.Data)
	}, workers, func(e Event[T]) {
		n := e.nonce
		eid := e.ID
		callbacks := r.callbacks
		handler(e.Data, func(data C, err error) {
			resp := Event[C]{
				ID:    eid,
				nonce: n + 1,
				Data:  data,
			}
			if err != nil {
				resp.Err = err
			}
			callbacks.Enqueue(resp)
		})
	})
}

func (r *reactor[T, C]) RemoveHandler(id string) {
	r.events.Unregister(id)
}

func (r *reactor[T, C]) AddCallback(id string, selector Selector[C], workers int, handler DemuxHandler[C]) {
	r.callbacks.Register(id, func(e Event[C]) bool {
		return selector(e.Data)
	}, workers, func(e Event[C]) {
		handler(e.Data)
	})
}

func (r *reactor[T, C]) RemoveCallback(id string) {
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
