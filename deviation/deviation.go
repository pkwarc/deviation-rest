package deviation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/montanaflynn/stats"
)

const (
	RandomApiUrl = "https://www.random.org/integers/"
	ApiLengthMax = 10000
	ApiLengthMin = 1
	TimeoutMs = 1000
)

type DevData struct {
	StdDev float64 `json:"stddev"`
	Data []float64   `json:"data"`
} 

type randomApiConfig struct {
	Num uint16
	format string
	min int
	max int
	col uint8
	base uint
}

func (config randomApiConfig) Url() string {
	return RandomApiUrl + 
		fmt.Sprintf("?num=%v", config.Num) +
		fmt.Sprintf("&format=%v", config.format) +
		fmt.Sprintf("&min=%v", config.min) +
		fmt.Sprintf("&max=%v", config.max) +
		fmt.Sprintf("&col=%v", config.col) +
		fmt.Sprintf("&base=%v", config.base)
}

func NewRandomApiConfig(num uint16) randomApiConfig {
	return randomApiConfig{
		Num: num,
		format: "plain",
		min: -1000000000,
		max: 1000000000,
		col: 1,
		base: 10,
	}
}

func RandomMeanHandler(w http.ResponseWriter, r *http.Request) {
	requests, err := strconv.ParseUint(r.URL.Query().Get("r"), 10, 32)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	length, err := strconv.ParseUint(r.URL.Query().Get("l"), 10, 32)
	if err != nil || length < ApiLengthMin || length > ApiLengthMax {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Fatalf("%v", err)
	}

	var counter uint64 = 0 
	ch := make(chan []float64)
	client := http.DefaultClient
	
	for counter < requests {
		counter++
		go func() {
			req, err := http.NewRequest("GET", NewRandomApiConfig(uint16(length)).Url(), nil)
			if err != nil {
				log.Fatalf("%v" ,err)
			}
			ctx, cancel := context.WithTimeout(req.Context(), TimeoutMs*time.Millisecond)
			defer cancel()
			req = req.WithContext(ctx)
			res, err := client.Do(req)
			ints, err := parseRandomApiResponse(res)
			ch <- ints
		}();
	}

	allNumbers := make([]float64, 0)
	numbers := make([][]float64, 0)
	for counter > 0 {
		counter--
		floats := <- ch 
		allNumbers = append(allNumbers, floats...)
		numbers = append(numbers, floats)
	}
	numbers = append(numbers, allNumbers)
	json.NewEncoder(w).Encode(CalculateDeviation(numbers))
}

func CalculateDeviation(data [][]float64) []DevData {
	results := make([]DevData, 0)
	for _, row := range data {
		fmt.Println(row)
		sdev, err := stats.StandardDeviation(row)
		fmt.Println(sdev)
		if err != nil {
			log.Printf("%v", err)
		} else {
			results = append(results, DevData{Data: row, StdDev: sdev})
		}

	}
	return results
}

func parseRandomApiResponse(res *http.Response) ([]float64, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	text := strings.TrimSpace(string(body)) 
	if res.StatusCode == http.StatusOK {
		elements := strings.Split(text, "\n")
		numbers := make([]float64, 0)
		for _, element := range elements {
			number, err := strconv.ParseFloat(element, 10)
			if err != nil {
				log.Printf("%v, cannot parse %v to int", err, element)
				continue
			}
			numbers = append(numbers, number)
		}
		return numbers, nil
	} else {
		return nil, errors.New(text)
	}
}
