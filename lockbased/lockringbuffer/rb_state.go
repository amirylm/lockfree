package lockringbuffer

type ringBufferState struct {
	head, tail uint32
	full       bool
}

// func newState(previous *ringBufferState) ringBufferState {
// 	if previous == nil {
// 		previous = new(ringBufferState)
// 	}
// 	return ringBufferState{
// 		head: previous.head,
// 		tail: previous.tail,
// 		full: previous.full,
// 	}
// }

func (state ringBufferState) Empty() bool {
	return !state.full && state.head == state.tail
}

func (state ringBufferState) Full() bool {
	return state.full
}
