package main

import (
	"github.com/spacemeshos/telemetry"
	"github.com/spacemeshos/telemetry/toolkit"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	c := telemetry.Channel{Name: "discovery", Origin: "generator", Proto: toolkit.Http{Endpoint: "http://localhost:9888/telemetry"}}.New()

	toolkit.ConstantCpuInfo("cpu", c)
	telemetry.ConstantString("sj.greeting", "hello!", c)

	telemetry.BindAutoString("additional.time", func() string {
		return time.Now().Format(time.Kitchen)
	}, c)

	t := time.NewTicker(time.Second * 1)
	e := make(chan os.Signal)
	signal.Notify(e, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

loop:
	for {
		select {
		case <-t.C:
			c.Commit()
		case <-e:
			break loop
		}
	}
}
