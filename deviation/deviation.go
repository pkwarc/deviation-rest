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
	RandomOrgUrl         = "https://www.random.org/integers/"
	RandomOrgLengthMax   = 10000
	RandomOrgLengthMin   = 1
	RandomOrgRequestsMin = 1
	RandomOrgRequestsMax = 100
	TimeoutMs            = 1000
)

var errLengthOutOfRange = fmt.Sprintf("'length' should be between %d and %d", RandomOrgLengthMin, RandomOrgLengthMax)
var errMaxRequestsOutOfRange = fmt.Sprintf("'requests' should be between %d and %d", RandomOrgRequestsMin, RandomOrgRequestsMax)

type DevData struct {
	StdDev float64   `json:"stddev"`
	Data   []float64 `json:"data"`
}

type ApiError struct {
	Err string
}

type GenerateNumbers func(context.Context, uint16) ([]float64, error)

func GetRandomMeanHandler(generate GenerateNumbers) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)

		requests, err := strconv.ParseUint(r.URL.Query().Get("requests"), 10, 32)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(ApiError{err.Error()})
			return
		}
		if requests < RandomOrgLengthMin || requests > RandomOrgRequestsMax {
			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(ApiError{errMaxRequestsOutOfRange})
			return
		}

		length, err := strconv.ParseUint(r.URL.Query().Get("length"), 10, 32)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(ApiError{err.Error()})
			return
		}
		if length < RandomOrgLengthMin || length > RandomOrgLengthMax {
			w.WriteHeader(http.StatusBadRequest)
			encoder.Encode(ApiError{errLengthOutOfRange})
			return
		}

		var counter uint64 = 0
		ch := make(chan []float64)
		cherr := make(chan error)
		ctx, cancel := context.WithTimeout(r.Context(), TimeoutMs*time.Millisecond)
		defer cancel()
		for counter < requests {
			counter++
			go func() {
				floats, err := generate(ctx, uint16(length))
				if err != nil {
					cherr <- err
					log.Printf("%v", err)
					return
				}
				ch <- floats
			}()
		}

		allNumbers := make([]float64, 0)
		numbers := make([][]float64, 0)
		var fail error
		for counter > 0 && fail == nil {
			counter--
			select {
			case floats := <-ch:
				allNumbers = append(allNumbers, floats...)
				numbers = append(numbers, floats)
			case err := <-cherr:
				fail = err
			}
		}

		if fail != nil {
			if errors.Is(fail, context.DeadlineExceeded) {
				w.WriteHeader(http.StatusRequestTimeout)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
			encoder.Encode(ApiError{fail.Error()})
			return
		}

		numbers = append(numbers, allNumbers)
		results := CalculateDeviation(numbers)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(results)
	})
}

func GetRandomOrgUrl(num uint16) string {
	return RandomOrgUrl +
		fmt.Sprintf("?num=%v", num) +
		fmt.Sprintf("&format=%v", "plain") +
		fmt.Sprintf("&min=%v", -1000000000) +
		fmt.Sprintf("&max=%v", 1000000000) +
		fmt.Sprintf("&col=%v", 1) +
		fmt.Sprintf("&base=%v", 10)
}

func randomGenerateNumbers(ctx context.Context, url string, num uint16) (data []float64, err error) {
	client := http.DefaultClient
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("%v", err)
	}
	req = req.WithContext(ctx)
	res, err := client.Do(req)
	if err != nil {
		return
	}
	return parseRandomApiResponse(res)
}

func RandomOrgGenerateNumbers(ctx context.Context, num uint16) (data []float64, err error) {
	url := GetRandomOrgUrl(num)
	return randomGenerateNumbers(ctx, url, num)
}

func CalculateDeviation(data [][]float64) []DevData {
	results := make([]DevData, 0)
	for _, row := range data {
		sdev, err := stats.StandardDeviation(row)
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
