package deviation

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func GenerateFixedNumbers(ctx context.Context, num uint16) ([]float64, error) {
	return []float64{1, 2, 3, 4, 5}, nil
}

func TestRandomMeanHandlerValidation(t *testing.T) {
	cases := []struct {
		Name string
		Method string
		Url string
		ExpectedStatus int
		ExpectedBody string
	} {
		{
			"Negative r and l",
			"GET",
			"/random/mean?r=-1&l=-1",
			400,
			`{"Err":"strconv.ParseUint: parsing \"-1\": invalid syntax"}`,
		},
		{
			"Positive r but l is missed",
			"GET",
			"/random/mean?r=1",
			400,
			`{"Err":"strconv.ParseUint: parsing \"\": invalid syntax"}`,
		},
		{
			"No r and l",
			"GET",
			"/random/mean",
			400,
			`{"Err":"strconv.ParseUint: parsing \"\": invalid syntax"}`,
		},
		{
			"l out of the valid range",
			"GET",
			"/random/mean?r=1&l=10001",
			400,
			`{"Err":"'l' should be between 1 and 10000"}`,
		},
		{
			"r out of the valid range",
			"GET",
			"/random/mean?r=1000&l=100",
			400,
			`{"Err":"'r' should be between 1 and 100"}`,
		},
	}

	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			req, err := http.NewRequest(test.Method, test.Url, nil)
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			handler := GetRandomMeanHandler(GenerateFixedNumbers)
			handler.ServeHTTP(recorder, req)
			body := strings.TrimSpace(recorder.Body.String())

			if status := recorder.Code; status != test.ExpectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, test.ExpectedStatus)
			}
			if body != test.ExpectedBody {
				t.Errorf("handler returned invalid body: got: %s want %s", body, test.ExpectedBody)
			}
		})
	}
}

func TestRandomMeanHandlerTimeout(t *testing.T) {
	slowServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second)
	}))
	slowGenerateNumbers := func(ctx context.Context, num uint16) ([]float64, error) {
		return randomGenerateNumbers(ctx, slowServer.URL, num)
	}
	req, err := http.NewRequest(http.MethodGet, "/random/mean?l=10&r=10", nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler := GetRandomMeanHandler(slowGenerateNumbers)
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusRequestTimeout {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusRequestTimeout)
	}
}

func TestRandomMeanHandlerStdDevResultIsOK(t *testing.T) {
	const l = 100
	const r = 10
	const randomInt = 49

	randomOrgMock := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		var builder strings.Builder
		counter := l
		// the response body in the random.org 1 col plain format
		for counter > 0 {
			counter--
			line := fmt.Sprintf("%d\n", randomInt)
			builder.WriteString(line)
		}
		rw.Write([]byte(builder.String()))
	}))
	generateNumbersMock := func(ctx context.Context, num uint16) ([]float64, error) {
		return randomGenerateNumbers(ctx, randomOrgMock.URL , num)
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/random/mean?l=%d&r=%d", l, r), nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler := GetRandomMeanHandler(generateNumbersMock)
	handler.ServeHTTP(recorder, req)

	var data []DevData
	err = json.Unmarshal(recorder.Body.Bytes(), &data) 
	if err != nil {
		t.Fatal(err)
	}

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	if len(data) != r + 1 {
		t.Errorf("got %v want %v", len(data), r + 1)
	} 
	const expectedStdDev = 0
	for idx := range data {
		result := data[idx]
		if result.StdDev != expectedStdDev {
			t.Errorf("got %v want %v", result.StdDev, expectedStdDev)
		}
		dataSetLen := len(result.Data)
		expectedLen := l
		if idx == r {
			expectedLen = l * r
		}
		if dataSetLen != expectedLen {
			t.Errorf("got %v want %v", dataSetLen, expectedLen)
		}
	}
}

func TestCalculateDeviation(t *testing.T) {
	input := [][]float64{
		{1, 2, 3, 4, 5},
		{1, 1, 2, 2, 3, 3, 4, 4, 5, 5},
		{41, 12, -33, 4, -15},
	}
	deviations := []float64{
		1.41421356,
		1.41421356,
		25.0551391,
	}

	results := CalculateDeviation(input)

	for idx, r := range results {
		got := r.Data
		want := input[idx]
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Got %v want %v", got, want)
		}
		devGot := big.NewFloat(r.StdDev).SetPrec(7)
		devWant := big.NewFloat(deviations[idx]).SetPrec(7)
		if devGot.Cmp(devWant) != 0 {
			t.Errorf("Got %f want %f", devGot, devWant)
		}

	}
}