package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
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
	// serves reader routines
	var wg1 sync.WaitGroup
	// serves writer routines
	var wg2 sync.WaitGroup
	// serves main routine for graceful termination
	var wg3 sync.WaitGroup
	wg1.Add(3)

	go read(c, 101, &wg1, &s, ds)
	go read(c, 202, &wg1, &s, ds)
	go read(c, 303, &wg1, &s, ds)

	wg3.Add(1)
	go func() {
		// fetch cryto-currency data from binance for processing
		res, err := http.Get("https://data.binance.com/api/v3/ticker/24hr")
		if err != nil {
			fmt.Println("Error making request: ", err)
			return
		}

		defer res.Body.Close()
		sc := bufio.NewScanner(res.Body)
		const max_c = 10 * 1024 * 1024 // 10MB (adjust as per your needs)
		b := make([]byte, max_c)
		sc.Buffer(b, max_c)
		sc.Split(scanJSON)

		for sc.Scan() {
			fmt.Println("Scanning input for next JSON entity")
			l := sc.Bytes()
			var tc examples.TickerData
			if err := json.Unmarshal(l, &tc); err != nil {
				log.Println("Error unmarshaling JSON:", err)
				continue // Skip malformed lines and proceed to the next one
			}
			wg2.Add(1)
			go writeData(c, tc, &wg2)
			time.Sleep(30 * time.Millisecond)
		}

		if err := sc.Err(); err != nil {
			panic(err)
		}

		// wait for writer routines to complete
		wg2.Wait()
		wg3.Done()
	}()
	wg3.Wait()
	s.v.Store(true)
	// wait for readers to complete
	wg1.Wait()
}

func writeData(c common.DataStructure[string], td examples.TickerData, wg *sync.WaitGroup) {
	// iterating over struct fields
	dv := reflect.ValueOf(&td).Elem()
	for i := 0; i < dv.NumField(); i++ {
		f := dv.Field(i)
		fn := string(dv.Type().Field(i).Name)
		fv := f.Interface()
		c.Push(fmt.Sprintf("%s: %v", fn, fv))
	}
	wg.Done()
}

func scanJSON(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Find the position of the next JSON object within the data
	s := strings.Index(string(data), "{")
	if s == -1 {
		// No opening brace found, request more data
		return 0, nil, nil
	}

	// Find the position of the corresponding closing brace '}' for the JSON object
	level := 1
	for i := s + 1; i < len(data); i++ {
		switch data[i] {
		case '{':
			level++
		case '}':
			level--
			if level == 0 {
				// Include the closing brace in the token
				return i + 1, data[s : i+1], nil
			}
		}
	}

	// JSON object is incomplete
	if atEOF {
		return 0, nil, err
	}

	// Request more data to complete the JSON object
	return 0, nil, nil
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
