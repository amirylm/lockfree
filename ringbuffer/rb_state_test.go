package ringbuffer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRingBufferState(t *testing.T) {
	require.Equal(t, uint64(0x80000000), newState(0).Uint64())
	require.True(t, newState(0).IsEmpty())
	require.Equal(t, uint64(0x80000001), newState(1).Uint64())
	state := &ringBufferState{
		head: 1,
		tail: 3,
		full: false,
	}
	require.False(t, state.IsEmpty())
	state.full = true
	state.head = uint32(12800)
	state.tail = uint32(12800)
	statecp := newState(state.Uint64())
	require.True(t, statecp.full)
	require.Equal(t, uint32(12800), statecp.head)
	require.Equal(t, uint32(12800), statecp.tail)
	statecp.head = uint32(4)
	statecp = newState(statecp.Uint64())
	require.Equal(t, uint32(4), statecp.head)
}
