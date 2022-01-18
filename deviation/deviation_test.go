package deviation

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"math/big"
)

func TestRandomMeanHandler(t *testing.T) {
	cases := []struct {
		Name string
		Method string
		Url string
		ExpectedStatus int
	} {
		{
			"Negative r and l",
			"GET",
			"/random/mean?r=-1&l=-1",
			400,
		},
		{
			"Positive r but l is missed",
			"GET",
			"/random/mean?r=-1",
			400,
		},
		{
			"No r and l",
			"GET",
			"/random/mean",
			400,
		},
		{
			"Positive r and l",
			"GET",
			"/random/mean?r=10&l=15",
			200,
		},
	}

	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			req, err := http.NewRequest(test.Method, test.Url, nil)
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(RandomMeanHandler)

			handler.ServeHTTP(recorder, req)

			if status := recorder.Code; status != test.ExpectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, test.ExpectedStatus)
			}
		})
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