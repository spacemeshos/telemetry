package tests

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/spacemeshos/telemetry"
	"github.com/spacemeshos/telemetry/toolkit"
	"gotest.tools/assert"
	"io"
	"io/ioutil"
	"net"
	"sync"
	"testing"
	"time"
)

type TelemetryTcpServer struct {
	l   net.Listener
	end chan struct{}
	wg  sync.WaitGroup
}

func StartTcpServer(t *testing.T, to *TestTelemetryOut, endpoint string) *TelemetryTcpServer {
	l, err := net.Listen("tcp", endpoint)
	if err != nil {
		t.Fatal(err.Error())
	}
	ts := &TelemetryTcpServer{l: l, end: make(chan struct{})}
	ts.wg.Add(1)
	go func() {
		defer ts.wg.Done()
		for {
			conn, err := ts.l.Accept()
			if err != nil {
				select {
				case <-ts.end:
					return
				default:
					t.Fatalf("tcp.Server.Accept(): %v", err)
					return
				}
			}
			b, err := ioutil.ReadAll(conn)
			if err != nil && err != io.EOF {
				t.Fatalf("tcp.Server.Read(): %v", err)
				return
			}
			sc := bufio.NewScanner(bytes.NewReader(b))
			for sc.Scan() {
				if len(sc.Bytes()) > 1 {
					pk := telemetry.Packet{}
					if err := json.NewDecoder(bytes.NewReader(sc.Bytes())).Decode(&pk); err != nil {
						t.Fatalf("tcp.Server.Read(): %v", err)
						return
					}
					bf := &bytes.Buffer{}
					encoder := json.NewEncoder(bf)
					encoder.SetIndent("", "\t")
					_ = encoder.Encode(pk)
					t.Logf("packet:\n%v", bf.String())
					_ = to.Send(pk)
				}
			}
		}
	}()
	return ts
}

func (ts *TelemetryTcpServer) Stop() {
	close(ts.end)
	_ = ts.l.Close()
	ts.wg.Wait()
}

func Test_TcpTelemetry(t *testing.T) {
	endpoint := "localhost:9888"
	to := &TestTelemetryOut{
		Channel: "ConstantStringTelemetry",
		Data: []J{
			{"string": "constant", "opt": J{"number": 0}},
			{"string": "constant", "opt": J{"number": 2}},
		},
	}

	s := StartTcpServer(t, to, endpoint)
	c := telemetry.Channel{
		Name:    to.Channel,
		Origin:  "",
		Proto:   toolkit.Auto{Endpoint: "tcp://" + endpoint, Timeout: 3 * time.Second},
		OnError: func(e error) { t.Error(e.Error()) },
	}.New()

	telemetry.ConstantString("string", "constant", c)
	n := telemetry.BindNumber("opt.number", nil, c)
	telemetry.Commit(c)
	n.Set(2)
	telemetry.Commit(c)
	telemetry.WaitAndClose(c)
	s.Stop()
	assert.Assert(t, to.Ok(), to.What())
}
