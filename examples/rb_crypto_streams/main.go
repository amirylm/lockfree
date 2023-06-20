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
	Volume             float64 `json:"volume,string"`
	OpenPrice          float64 `json:"openPrice,string"`
	HighPrice          float64 `json:"highPrice,string"`
	LowPrice           float64 `json:"lowPrice,string"`
}

func main() {
	stream := make(chan string)
	rb := ringbuffer.New[string](128)
	// buffer := make(chan string, 10)

	var wg sync.WaitGroup
	wg.Add(3)

	go readStream(stream, rb, &wg)
	go readBuffer(rb, 101, &wg)
	go readBuffer(rb, 202, &wg)

	go func() {
		res, err := http.Get("https://data.binance.com/api/v3/ticker/24hr")
		if err != nil {
			fmt.Println("Error making request: ", err)
			return
		}
		defer res.Body.Close()

		// read body
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
				// fmt.Printf("%s: %v\n", fieldName, fieldValue)
				stream <- fmt.Sprintf("%s: %v", fieldName, fieldValue)
			}
			time.Sleep(250 * time.Millisecond)
		}
	}()
	wg.Wait()
}

func readStream(stream <-chan string, rb common.DataStructure[string], wg *sync.WaitGroup) {
	defer wg.Done()
	for message := range stream {
		rb.Push(message)
	}
}

func readBuffer(rb common.DataStructure[string], rid int, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		if !rb.Empty() {
			v, succeeded := rb.Pop()
			if succeeded {
				fmt.Printf("From %d : %v\n", rid, v)
			}
		}
		time.Sleep(400 * time.Millisecond)
	}
}
