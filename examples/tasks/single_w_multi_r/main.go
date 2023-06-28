package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/amirylm/lockfree/common"
	"github.com/amirylm/lockfree/queue"
	"github.com/amirylm/lockfree/ringbuffer"
	"github.com/amirylm/lockfree/stack"
)

type Task struct {
	ID   int
	Data string
}

type State struct {
	v atomic.Bool
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go ringbuffer|queue|stack")
		return
	}
	ds := os.Args[1]

	var c common.DataStructure[Task]
	switch ds {
	case "ringbuffer":
		c = ringbuffer.New[Task](1001)
	case "queue":
		c = queue.New[Task](1001)
	case "stack":
		c = stack.New[Task](1001)
	default:
		fmt.Println("Illegal argument. Usage ringbuffer | queue | stack")
		return
	}

	s := State{}
	s.v.Store(false)
	var wg1 sync.WaitGroup
	var wg2 sync.WaitGroup

	// Create multiple workers to process tasks concurrently
	workers := 5
	for i := 0; i < workers; i++ {
		wg1.Add(1)
		go processTasks(i, c, &wg1, &s, ds)
	}

	wg2.Add(1)
	// Producer: Add tasks to the queue
	go func() {
		for i := 1; i <= 1000; i++ {
			c.Push(Task{ID: i, Data: fmt.Sprintf("Task %d", i)})
		}
		fmt.Println("All tasks populated")
		fmt.Printf("Data structure contains %d\n", c.Size())
		wg2.Done()
	}()

	wg2.Wait()
	s.v.Store(true)
	wg1.Wait()
	fmt.Println("Processing of all tasks has been completed")
}

func processTasks(workerId int, c common.DataStructure[Task], wg *sync.WaitGroup, s *State, ds string) {
	for {
		if !c.Empty() {
			task, ok := c.Pop()
			if ok {
				// Process the task
				fmt.Printf("Worker %d processing task: %d\n", workerId, task.ID)
			}
		}

		if c.Empty() && s.v.Load() {
			fmt.Printf("From %d : %s is empty and state of population is done, Terminating gracefully.\n", workerId, ds)
			break
		}
		runtime.Gosched()
	}
	wg.Done()
}
