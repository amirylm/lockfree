package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/amirylm/lockfree/core"
	"github.com/amirylm/lockfree/examples/streams"
)

func main() {
	var c core.Queue[string]
	var ds string
	args := os.Args
	c, ds = streams.PromptDS(args)

	s := streams.State{}
	s.SetValue(false)
	// serves reader routines
	var wg1 sync.WaitGroup
	// serves writer routines
	var wg2 sync.WaitGroup
	// serves main routine for graceful termination
	var wg3 sync.WaitGroup
	wg1.Add(3)

	go streams.Read(c, 101, &wg1, &s, ds)
	go streams.Read(c, 202, &wg1, &s, ds)
	go streams.Read(c, 303, &wg1, &s, ds)

	wg3.Add(1)
	go func() {
		// fetch cryto-currency data from binance for processing
		res, err := http.Get("https://data.binance.com/api/v3/ticker/24hr")
		if err != nil {
			fmt.Println("Error making request: ", err)
			return
		}

		defer func() {
			if err := res.Body.Close(); err != nil {
				fmt.Println("Error closing body:", err)
			}
		}()
		decoder := json.NewDecoder(res.Body)
		var tc []streams.TickerData
		err = decoder.Decode(&tc)

		if err != nil {
			panic("Error decoding data")
		}

		for _, t := range tc {
			wg2.Add(1)
			go streams.WriteData(c, t, &wg2)
			time.Sleep(30 * time.Millisecond)
		}

		// wait for writer routines to complete
		wg2.Wait()
		wg3.Done()
	}()
	wg3.Wait()
	s.SetValue(true)
	// wait for readers to complete
	wg1.Wait()
}
