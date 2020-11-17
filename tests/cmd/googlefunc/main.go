package main

import (
	fw "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"github.com/spacemeshos/telemetry/toolkit"
	"log"
	"os"
)

func main() {
	_ = os.Setenv("TELEMETRY_VERBOSE", "1")
	fw.RegisterHTTPFunction("/telemetry", toolkit.TelemetryFunction())
	if err := fw.Start("9888"); err != nil {
		log.Fatalf("framework.Start: %v\n", err)
	}
}
