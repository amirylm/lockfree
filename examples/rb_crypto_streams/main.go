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

func main() {
	stream := make(chan string)
	rb := ringbuffer.New[string](128)

	var wg sync.WaitGroup
	wg.Add(3)

	go readBuffer(rb, 101, &wg)
	go readBuffer(rb, 202, &wg)
	go readBuffer(rb, 303, &wg)

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

		var tickerData []TickerData
		err = json.Unmarshal(body, &tickerData)

		if err != nil {
			fmt.Println("Error unmarshaling JSON: ", err)
			return
		}

		defer close(stream)
		for i := 0; i < len(tickerData); i++ {
			dataValue := reflect.ValueOf(&tickerData[i]).Elem()
			for i := 0; i < dataValue.NumField(); i++ {
				field := dataValue.Field(i)
				fieldName := string(dataValue.Type().Field(i).Name)
				fieldValue := field.Interface()
				rb.Push(fmt.Sprintf("%s: %v", fieldName, fieldValue))
			}
			time.Sleep(250 * time.Millisecond)
		}
	}()
	wg.Wait()
}

func readBuffer(rb common.DataStructure[string], rid int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		if !rb.Empty() {
			v, ok := rb.Pop()
			if ok {
				fmt.Printf("From %d : %v\n", rid, v)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}
