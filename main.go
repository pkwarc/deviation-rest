package main

import (
	"net/http"
	"github.com/pkwarc/deviation-rest/deviation"
)

func main() {
	numbersGenerator := deviation.RandomOrgGenerateNumbers
	handler := deviation.GetRandomMeanHandler(numbersGenerator)
	http.Handle("/random/mean", handler)
	http.ListenAndServe(":7777", nil)
}
