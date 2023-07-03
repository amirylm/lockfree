package streams

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/amirylm/lockfree/core"
)

type State struct {
	v atomic.Bool
}

func (s *State) SetValue(value bool) {
	s.v.Store(value)
}

func (s *State) GetValue() bool {
	return s.v.Load()
}

func WriteData(c core.Queue[string], td TickerData, wg *sync.WaitGroup) {
	// iterating over struct fields
	dv := reflect.ValueOf(&td).Elem()
	for i := 0; i < dv.NumField(); i++ {
		f := dv.Field(i)
		fn := string(dv.Type().Field(i).Name)
		fv := f.Interface()
		c.Enqueue(fmt.Sprintf("%s: %v", fn, fv))
	}
	wg.Done()
}

func Read(c core.Queue[string], rid int, wg *sync.WaitGroup, s *State, ds string) {
	for {
		if c.Empty() && s.v.Load() {
			fmt.Printf("From %d : %s is empty and state of population is done, Terminating gracefully.\n", rid, ds)
			break
		}
		if !c.Empty() {
			v, ok := c.Dequeue()
			if ok {
				fmt.Printf("From %d : %v\n", rid, v)
			}
		}
	}
	wg.Done()
	time.Sleep(30 * time.Millisecond)
}
