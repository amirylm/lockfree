package reactor

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/amirylm/go-options"
)

type Encoder[T any] interface {
	Clone(T) T
	ID(T) []byte
}

type Reactor[T any] interface {
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

func WithEventsDemux[T any](d Demultiplexer[event[T]]) options.Option[reactor[T]] {
	return func(r *reactor[T]) {
		r.events = d
	}
}

func WithCallbacksDemux[T any](d Demultiplexer[event[T]]) options.Option[reactor[T]] {
	return func(r *reactor[T]) {
		r.callbacks = d
	}
}

func WithEncoder[T any](encoder Encoder[T]) options.Option[reactor[T]] {
	return func(r *reactor[T]) {
		r.encoder = encoder
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
		r.events = NewDemux[event[T]]()
	}
	if r.callbacks == nil {
		r.callbacks = NewDemux[event[T]]()
	}

	return r
}

type reactor[T any] struct {
	events, callbacks Demultiplexer[event[T]]
	tick, timeout     time.Duration

	encoder Encoder[T]
}

func (r *reactor[T]) genID(T) ID {
	return []byte(fmt.Sprintf("%04d-%08d-%04d",
		rand.Intn(9999), rand.Intn(99999999), rand.Intn(9999)))
}

func (r *reactor[T]) Enqueue(events ...T) {
	for _, data := range events {
		r.events.Enqueue(event[T]{
			id:    r.genID(data),
			nonce: 0,
			data:  data,
		})
	}
}

func (r *reactor[T]) EnqueueWait(pctx context.Context, data T) (T, error) {
	ctx, cancel := context.WithTimeout(pctx, r.timeout)
	defer cancel()

	resultp := atomic.Pointer[event[T]]{}
	nonce := int64(1)
	id := r.genID(data)

	cid := fmt.Sprintf("%x:%d", id, nonce)
	r.callbacks.Register(cid, func(e event[T]) bool {
		return bytes.Equal(e.id, id) && e.nonce == nonce
	}, 0, func(e event[T]) {
		resultp.Store(&e)
	})
	defer r.callbacks.Unregister(cid)

	r.events.Enqueue(event[T]{
		id:    id,
		nonce: nonce - 1,
		data:  data,
	})

	result := resultp.Load()
	for ctx.Err() == nil && result == nil {
		time.Sleep(r.tick)
		result = resultp.Load()
	}

	if result != nil {
		return result.data, result.err
	}
	var res T
	return res, ctx.Err()
}

func (r *reactor[T]) AddHandler(id string, selector Selector[T], workers int, handler EventHandler[T]) {
	r.events.Register(id, func(e event[T]) bool {
		return selector(e.data)
	}, workers, func(e event[T]) {
		n := e.nonce
		eid := e.id
		callbacks := r.callbacks
		handler(e.data, func(data T, err error) {
			resp := event[T]{
				id:    eid,
				nonce: n + 1,
				data:  data,
			}
			if err != nil {
				resp.err = err
			}
			callbacks.Enqueue(resp)
		})
	})
}

func (r *reactor[T]) RemoveHandler(id string) {
	r.events.Unregister(id)
}

func (r *reactor[T]) AddCallback(id string, selector Selector[T], workers int, handler DemuxHandler[T]) {
	r.callbacks.Register(id, func(e event[T]) bool {
		return selector(e.data)
	}, workers, func(e event[T]) {
		handler(e.data)
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

type event[T any] struct {
	id    ID
	nonce int64
	data  T
	err   error
}
