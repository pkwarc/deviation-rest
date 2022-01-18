package main

import (
	"net/http"
	"github.com/pkwarc/deviation-rest/deviation"
)


func main() {
	http.HandleFunc("/random/mean", deviation.RandomMeanHandler)
	http.ListenAndServe(":5051", nil)
}
