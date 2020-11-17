package main

import (
	"context"
	"github.com/spacemeshos/telemetry/toolkit"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {

	ts := StartServer("http://localhost:9888/telemetry", "localhost:9898")

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-c

	ts.Stop()
}

type TelemetryServer struct {
	http.Server
	toolkit.Reporter
	wg sync.WaitGroup
}

func StartServer(gwEndpoint, lgsEndpoint string) *TelemetryServer {
	u, _ := url.Parse(gwEndpoint)
	r := toolkit.Reporter{toolkit.Tcp{Endpoint: lgsEndpoint}.New(), log.Printf, nil}
	ts := &TelemetryServer{Server: http.Server{Addr: u.Host, Handler: http.HandlerFunc(r.Report)}, Reporter: r}
	ts.wg.Add(1)
	go func() {
		defer ts.wg.Done()
		if err := ts.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("http.TelemetryServer.ListenAndServe(): %v", err)
		}
		ts.Reporter.Close()
	}()
	return ts
}

func (ts *TelemetryServer) Stop() {
	if err := ts.Shutdown(context.TODO()); err != nil {
		log.Fatalf("http.TelemetryServer.Shutdown(): %v", err)
	}
	ts.wg.Wait()
}
