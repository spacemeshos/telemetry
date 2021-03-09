package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/spacemeshos/telemetry"
	"github.com/spacemeshos/telemetry/toolkit"
	"gotest.tools/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

type TelemetrySendWrapper func([]byte) error

func (w TelemetrySendWrapper) Send(pk telemetry.Packet) error {
	js, err := json.Marshal(pk)
	if err != nil {
		return err
	}
	return w(js)
}

func (w TelemetrySendWrapper) Close() error {
	return nil
}

type TelemetryHttpServer struct {
	http.Server
	toolkit.Reporter
	wg sync.WaitGroup
}

func StartHttpServer(t *testing.T, to *TestTelemetryOut, endpoint string, validate func(pk telemetry.Packet) error) *TelemetryHttpServer {
	u, _ := url.Parse(endpoint)
	r := toolkit.Reporter{TelemetrySendWrapper(to.SenderImpl), t.Logf, validate}
	ts := &TelemetryHttpServer{Server: http.Server{Addr: fmt.Sprintf("127.0.0.1:%s", u.Port()), Handler: http.HandlerFunc(r.Report)}, Reporter: r}
	ts.wg.Add(1)
	go func() {
		defer ts.wg.Done()
		if err := ts.Server.ListenAndServe(); err != http.ErrServerClosed {
			t.Fatalf("http.Server.ListenAndServe(): %v", err)
		}
		ts.Reporter.Close()
	}()
	return ts
}

func (ts *TelemetryHttpServer) Stop(t *testing.T) {
	if err := ts.Server.Shutdown(context.TODO()); err != nil {
		t.Fatalf("http.Server.Shutdown(): %v", err)
	}
	ts.wg.Wait()
}

func Test_ConstantStringTelemetry(t *testing.T) {
	to := &TestTelemetryOut{
		Channel: "ConstantStringTelemetry",
		Data: []J{
			{"string": "constant"},
		}}

	to.Sender = func(js []byte) error {
		req, err := http.NewRequest("POST", "/telemetry", bytes.NewBuffer(js))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		toolkit.Reporter{TelemetrySendWrapper(to.SenderImpl), t.Logf, nil}.Report(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
		rs := toolkit.HttpResult{}
		if err = json.Unmarshal(rr.Body.Bytes(), &rs); err != nil {
			t.Fatal(err)
		}
		if rs.Status != "Ok" {
			t.Errorf("report function failed: %v", rs.Error)
		}
		return nil
	}

	c := telemetry.Channel{to.Channel, "", to, func(e error) { t.Error(e.Error()) }}.New()
	telemetry.ConstantString("string", "constant", c)
	c.Commit()
	telemetry.WaitAndClose(c)
	assert.Assert(t, to.Ok(), to.What())
}

func doPublicTelemetry(t *testing.T, compressed bool) {
	endpoint := "http://localhost:9888/telemetry"
	to := &TestTelemetryOut{
		Channel: "ConstantStringTelemetry",
		Data: []J{
			{"string": "constant", "opt": J{"number": 0}},
			{"string": "constant", "opt": J{"number": 2}},
		},
	}

	s := StartHttpServer(t, to, endpoint, func(pk telemetry.Packet) error {
		ok := compressed == (len(pk.Telemetry.Compressed) != 0) ||
			len(pk.Telemetry.Data) == 0 && len(pk.Telemetry.Compressed) == 0
		if !ok {
			s := ""
			if compressed {
				s = "not "
			}
			return fmt.Errorf("packet is %scompressed", s)
		}
		return nil
	})
	defer s.Stop(t)

	c := telemetry.Channel{
		to.Channel, "",
		toolkit.Auto{Endpoint: endpoint, Timeout: 3 * time.Second, Compressed: compressed},
		func(e error) { t.Error(e.Error()) }}.New()

	telemetry.ConstantString("string", "constant", c)
	n := telemetry.BindNumber("opt.number", nil, c)
	c.Commit()
	n.Set(2)
	c.Commit()
	telemetry.WaitAndClose(c)
	assert.Assert(t, to.Ok(), to.What())
}

func Test_CompressedPublicTelemetry(t *testing.T) {
	doPublicTelemetry(t, true)
}

func Test_UncompressedPublicTelemetry(t *testing.T) {
	doPublicTelemetry(t, false)
}
