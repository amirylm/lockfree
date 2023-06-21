package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
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
	v atomic.Bool
}

func main() {
	rb := ringbuffer.New[string](128)
	done := State{}
	done.v.Store(false)
	// serves reader routines
	var wg1 sync.WaitGroup
	// serves writer routines
	var wg2 sync.WaitGroup
	// serves main routine for graceful termination
	var wg3 sync.WaitGroup
	wg1.Add(3)

	go readBuffer(rb, 101, &wg1, &done)
	go readBuffer(rb, 202, &wg1, &done)
	go readBuffer(rb, 303, &wg1, &done)

	wg3.Add(1)
	go func() {
		// fetch cryto-currency data from binance for processing
		res, err := http.Get("https://data.binance.com/api/v3/ticker/24hr")
		if err != nil {
			fmt.Println("Error making request: ", err)
			return
		}

		defer res.Body.Close()
		scanner := bufio.NewScanner(res.Body)
		const maxCapacity = 10 * 1024 * 1024 // 10MB (adjust as per your needs)
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)
		scanner.Split(scanConcatenatedJSON)

		for scanner.Scan() {
			fmt.Println("Scanning input for next JSON entity")
			line := scanner.Bytes()
			var data TickerData
			if err := json.Unmarshal(line, &data); err != nil {
				log.Println("Error unmarshaling JSON:", err)
				continue // Skip malformed lines and proceed to the next one
			}
			wg2.Add(1)
			go writeTickerDataToBuffer(rb, data, &wg2)
			time.Sleep(30 * time.Millisecond)
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		// wait for writer routines to complete
		wg2.Wait()
		// trigger state change signaling writers have completed
		wg3.Done()
	}()
	wg3.Wait()
	done.v.Store(true)
	// wait for readers to complete
	wg1.Wait()
}

func writeTickerDataToBuffer(rb common.DataStructure[string], td TickerData, wg *sync.WaitGroup) {
	// iterating over struct fields
	dataV := reflect.ValueOf(&td).Elem()
	for i := 0; i < dataV.NumField(); i++ {
		field := dataV.Field(i)
		fieldN := string(dataV.Type().Field(i).Name)
		fieldV := field.Interface()
		rb.Push(fmt.Sprintf("%s: %v", fieldN, fieldV))
	}
	wg.Done()
}

func scanConcatenatedJSON(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Find the position of the next JSON object within the data
	start := strings.Index(string(data), "{")
	if start == -1 {
		// No opening brace found, request more data
		return 0, nil, nil
	}

	// Find the position of the corresponding closing brace '}' for the JSON object
	level := 1
	for i := start + 1; i < len(data); i++ {
		switch data[i] {
		case '{':
			level++
		case '}':
			level--
			if level == 0 {
				// Include the closing brace in the token
				return i + 1, data[start : i+1], nil
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

func readBuffer(rb common.DataStructure[string], rid int, wg *sync.WaitGroup, s *State) {
	for {
		done := s.v.Load()
		if !rb.Empty() {
			v, ok := rb.Pop()
			if ok {
				fmt.Printf("From %d : %v\n", rid, v)
			}
		}
		if rb.Empty() && done {
			fmt.Printf("From %d : Ringbuffer is empty and state of population is done, Terminating gracefully.\n", rid)
			break
		}
		runtime.Gosched()
	}
	wg.Done()
}
