package core

import (
	"sync/atomic"

	"github.com/amirylm/go-options"
)

// Options is the configuration for data structures (stacks or queues)
type Options struct {
	// capacity is the max size of the data structure
	capacity atomic.Int32
	// override is a flag that determines whether the data source will allow overriding records or not.
	// NOTE: applicable only for ring buffer
	override bool
}

// Capacity returns the capacity config, thread safe
func (o *Options) Capacity() int32 {
	return o.capacity.Load()
}

func (o *Options) Override() bool {
	return o.override
}

func WithCapacity(c int) options.Option[Options] {
	return func(opts *Options) {
		opts.capacity.Store(int32(c))
	}
}

func WithOverride(o bool) options.Option[Options] {
	return func(opts *Options) {
		opts.override = o
	}
}
