package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/amirylm/lockfree/common"
	examples "github.com/amirylm/lockfree/examples/streams"
	"github.com/amirylm/lockfree/queue"
	"github.com/amirylm/lockfree/ringbuffer"
	"github.com/amirylm/lockfree/stack"
)

type State struct {
	v atomic.Bool
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go ringbuffer|queue|stack")
		return
	}
	ds := os.Args[1]

	var c common.DataStructure[string]
	switch ds {
	case "ringbuffer":
		c = ringbuffer.New[string](128)
	case "queue":
		c = queue.New[string](128)
	case "stack":
		c = stack.New[string](128)
	default:
		fmt.Println("Illegal argument. Must be ringbuffer | queue | stack")
		return
	}

	s := State{}
	s.v.Store(false)
	var wg sync.WaitGroup
	wg.Add(3)

	go read(c, 101, &wg, &s, ds)
	go read(c, 202, &wg, &s, ds)
	go read(c, 303, &wg, &s, ds)

	go func() {
		// fetch crytp-currency data from binance for processing
		res, err := http.Get("https://data.binance.com/api/v3/ticker/24hr")
		if err != nil {
			fmt.Println("Error making request: ", err)
			return
		}

		defer res.Body.Close()
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("Error reading response body: ", err)
			return
		}

		var td []examples.TickerData
		err = json.Unmarshal(b, &td)

		if err != nil {
			fmt.Println("Error unmarshaling JSON: ", err)
			return
		}

		for i := 0; i < len(td); i++ {

			// iterating over struct fields
			dv := reflect.ValueOf(&td[i]).Elem()
			for i := 0; i < dv.NumField(); i++ {
				f := dv.Field(i)
				fn := string(dv.Type().Field(i).Name)
				fv := f.Interface()
				c.Push(fmt.Sprintf("%s: %v", fn, fv))
			}
			// simulating case where writing to data structure (21 data-points per iteration)
			// is at faster velocity than reading - thus requiring multiple reader routines.
			time.Sleep(30 * time.Millisecond)
		}
		s.v.Store(true)
	}()
	wg.Wait()
}

func read(c common.DataStructure[string], rid int, wg *sync.WaitGroup, s *State, ds string) {
	for {
		if !c.Empty() {
			v, ok := c.Pop()
			if ok {
				fmt.Printf("From %d : %v\n", rid, v)
			}
		}
		if c.Empty() && s.v.Load() {
			fmt.Printf("From %d : %s is empty and state of population is done, Terminating gracefully.\n", rid, ds)
			break
		}
		runtime.Gosched()
	}
	wg.Done()
}
