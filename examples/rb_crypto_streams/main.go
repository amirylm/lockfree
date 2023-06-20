package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/amirylm/lockfree/common"
	"github.com/amirylm/lockfree/ringbuffer"
)

type TickerData struct {
	Symbol             string  `json:"symbol"`
	PriceChange        float64 `json:"priceChange,string"`
	PriceChangePercent float64 `json:"priceChangePercent,string"`
	WeightedAvgPrice   float64 `json:"weightedAvgPrice,string"`
	PrevClosePrice     float64 `json:"prevClosePrice,string"`
	LastPrice          float64 `json:"lastPrice,string"`
	LastQty            float64 `json:"lastQty,string"`
	BidPrice           float64 `json:"bidPrice,string"`
	BidQty             float64 `json:"bidQty,string"`
	AskPrice           float64 `json:"askPrice,string"`
	AskQty             float64 `json:"askQty,string"`
	Volume             float64 `json:"volume,string"`
	OpenPrice          float64 `json:"openPrice,string"`
	HighPrice          float64 `json:"highPrice,string"`
	LowPrice           float64 `json:"lowPrice,string"`
	QuoteVolume        float64 `json:"quoteVolume,string"`
	OpenTime           int64   `json:"openTime"`
	CloseTime          int64   `json:"closeTime"`
	FirstId            int64   `json:"firstId"`
	LastId             int64   `json:"lastId"`
	Count              int64   `json:"count"`
}

type State struct {
	v  bool
	mu sync.RWMutex
}

func main() {
	rb := ringbuffer.New[string](128)
	done := State{v: false}
	var wg sync.WaitGroup
	wg.Add(3)

	go readBuffer(rb, 101, &wg, &done)
	go readBuffer(rb, 202, &wg, &done)
	go readBuffer(rb, 303, &wg, &done)

	go func() {
		// fetch crytp-currency data from binance for processing
		res, err := http.Get("https://data.binance.com/api/v3/ticker/24hr")
		if err != nil {
			fmt.Println("Error making request: ", err)
			return
		}

		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println("Error reading response body: ", err)
			return
		}

		var td []TickerData
		err = json.Unmarshal(body, &td)

		if err != nil {
			fmt.Println("Error unmarshaling JSON: ", err)
			return
		}

		for i := 0; i < len(td); i++ {

			// iterating over struct fields
			dataV := reflect.ValueOf(&td[i]).Elem()
			for i := 0; i < dataV.NumField(); i++ {
				field := dataV.Field(i)
				fieldN := string(dataV.Type().Field(i).Name)
				fieldV := field.Interface()
				rb.Push(fmt.Sprintf("%s: %v", fieldN, fieldV))
			}
			time.Sleep(30 * time.Millisecond)
		}
		done.mu.Lock()
		done.v = true
		done.mu.Unlock()
	}()
	wg.Wait()
}

func readBuffer(rb common.DataStructure[string], rid int, wg *sync.WaitGroup, s *State) {
	// defer wg.Done()
	for {
		s.mu.RLock()
		done := s.v
		s.mu.RUnlock()
		if !rb.Empty() {
			v, ok := rb.Pop()
			if ok {
				fmt.Printf("From %d : %v\n", rid, v)
			}
		}
		if rb.Empty() && done {
			fmt.Printf("From %d : Ringbuffer is empty and state of population is done, Terminating gracefully.", rid)
			break
		}
		time.Sleep(4 * time.Millisecond)
	}
	wg.Done()
}
