package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
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
	var wg sync.WaitGroup
	var wg2 sync.WaitGroup
	wg.Add(3)

	go streams.Read(c, 101, &wg, &s, ds)
	go streams.Read(c, 202, &wg, &s, ds)
	go streams.Read(c, 303, &wg, &s, ds)

	go func() {
		// fetch crytp-currency data from binance for processing
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
		b, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println("Error reading response body: ", err)
			return
		}

		var tda []streams.TickerData
		err = json.Unmarshal(b, &tda)

		if err != nil {
			fmt.Println("Error unmarshaling JSON: ", err)
			return
		}

		for i := 0; i < len(tda); i++ {
			td := tda[i]
			wg2.Add(1)
			go streams.WriteData(c, td, &wg2)
			// simulating case where writing to data structure (21 data-points per iteration)
			// is at faster velocity than reading - thus requiring multiple reader routines.
			time.Sleep(30 * time.Millisecond)
			runtime.Gosched()
		}
		wg2.Wait()
		s.SetValue(true)
	}()
	wg.Wait()
}
