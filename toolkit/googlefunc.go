package toolkit

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func endpoint()string {
	ep, ok := os.LookupEnv("TELEMETRY_ENDPOINT")
	if !ok {
		ep = "localhost:9898"
	}
	return ep
}

func timeout() time.Duration {
	tm, ok := os.LookupEnv("TELEMETRY_TIMEOUT")
	if !ok {
		tm = "3"
	}
	i, err := strconv.Atoi(tm)
	if err != nil {
		i = 3
	}
	return time.Duration(i)*time.Second
}

func verbose() func(string,...interface{}){
	v, ok := os.LookupEnv("TELEMETRY_VERBOSE")
	if !ok || v == "" || v == "0" {
		return nil
	}
	return log.Printf
}

func TelemetryFunction() func(w http.ResponseWriter, r *http.Request) {
	return Reporter{
		Protocol: Tcp{Endpoint: endpoint(), Timeout: timeout()}.New(),
		Verbose: verbose(),
	}.Report
}
