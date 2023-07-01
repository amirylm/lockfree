package ringbuffer

// ringBufferState holds the state of the ring buffer
// the state can be de/encoded to uin64 to be stored as an atomic.Uint64.
type ringBufferState struct {
	head, tail uint32
	full       bool
}

func newState(state uint64) ringBufferState {
	full := (state & 0x40000000) != 0
	head := uint32((state >> 32) & 0x0000FFFF)
	tail := uint32(state & 0x0000FFFF)
	return ringBufferState{
		head: head,
		tail: tail,
		full: full,
	}
}

func (state ringBufferState) Uint64() uint64 {
	fullBit := uint64(0)
	if state.full {
		fullBit = 0x40000000
	}

	headBits := uint64(state.head) << 32
	tailBits := uint64(state.tail)

	return 0x80000000 | fullBit | headBits | tailBits
}

func (state ringBufferState) Full() bool {
	return state.full
}

func (state ringBufferState) Empty() bool {
	return !state.full && state.head == state.tail
}
