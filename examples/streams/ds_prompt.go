package streams

import (
	"github.com/amirylm/lockfree/common"
	"github.com/amirylm/lockfree/queue"
	"github.com/amirylm/lockfree/ringbuffer"
	"github.com/amirylm/lockfree/stack"
)

func PromptDS(args []string) (common.DataStructure[string], string) {
	if len(args) < 2 {
		panic("Usage: go run main.go ringbuffer|queue|stack")
	}
	ds := args[1]

	var c common.DataStructure[string]
	switch ds {
	case "ringbuffer":
		c = ringbuffer.New[string](128)
	case "queue":
		c = queue.New[string](100000)
	case "stack":
		c = stack.New[string](128)
	default:
		panic("Illegal argument. Must be ringbuffer | queue | stack")
	}
	return c, ds
}
