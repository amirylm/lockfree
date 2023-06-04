package ringbuffer

type ringBufferState struct {
	head, tail uint32
	full       bool
}

func newState(prev *ringBufferState) *ringBufferState {
	state := new(ringBufferState)
	if prev != nil {
		state.full = prev.full
		state.head = prev.head
		state.tail = prev.tail
	}
	return state
}

func (state *ringBufferState) SetFull(full bool) {
	state.full = full
}

func (state *ringBufferState) BumpTail() uint32 {
	state.tail++
	return state.tail
}

func (state *ringBufferState) BumpHead() uint32 {
	state.head++
	return state.head
}

func (state *ringBufferState) Tail() uint32 {
	return state.tail
}

func (state *ringBufferState) Head() uint32 {
	return state.head
}

func (state *ringBufferState) Full() bool {
	return state.full
}

func (state *ringBufferState) Empty() bool {
	return !state.full && state.head == state.tail
}
