package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"fmt"

	"github.com/pkwarc/deviation-rest/deviation"
)

const (
	DEFAULT_PORT = 7777
	PORT_ENV_NAME = "STD_DEV_PORT"
)

func main() {
	port := GetEnvPort(PORT_ENV_NAME, DEFAULT_PORT)
	numbersGenerator := deviation.RandomOrgGenerateNumbers
	handler := deviation.GetRandomMeanHandler(numbersGenerator)
	http.Handle("/random/mean", handler)
	address := fmt.Sprintf(":%d", port)
	log.Println("The api available at " + address)
	log.Fatal(http.ListenAndServe(address, nil))
}

func GetEnvPort(envName string, fallbackPort uint16) uint16 {
	var port uint16 = fallbackPort
	envPort := os.Getenv(envName)
	if envPort != "" {
		userPort, err := strconv.ParseUint(envPort, 10, 16)
		if err != nil {
			log.Fatal(envName + " must be a valid port number")
		}
		port = uint16(userPort)
	}
	return port
}