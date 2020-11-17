package toolkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spacemeshos/telemetry"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type protoHttp struct {
	http       *http.Client
	endpoint   string
	compressed bool
	mu         sync.Mutex
}

type Http struct {
	Endpoint   string
	Timeout    time.Duration
	Compressed bool
}

func (c Http) New() telemetry.Protocol {
	return &protoHttp{
		http:       &http.Client{Timeout: c.Timeout},
		endpoint:   c.Endpoint,
		compressed: c.Compressed,
	}

}

type HttpResult struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func (pt *protoHttp) Send(pk telemetry.Packet) error {

	pt.mu.Lock()
	ht := pt.http
	pt.mu.Unlock()

	if ht == nil {
		return fmt.Errorf("http protocol is closed")
	}

	if pt.compressed {
		val, err := Compress(pk.Telemetry.Data)
		if err != nil {
		}
		pk.Telemetry.Compressed = val
		pk.Telemetry.Data = nil
	}
	js, err := json.Marshal(pk)
	if err != nil {
		return err
	}
	rq, err := http.NewRequest(http.MethodPost, pt.endpoint, bytes.NewBuffer(js))
	rq.Header.Set("Content-Type", "application/json; charset=utf-8")

	rp, err := ht.Do(rq)
	if err != nil {
		return err
	}
	defer rp.Body.Close()
	if rp.StatusCode != 200 {
		return fmt.Errorf("http error: %d %v", rp.StatusCode, rp.StatusCode)
	}
	body, err := ioutil.ReadAll(rp.Body)
	if err != nil {
		return err
	}
	rs := HttpResult{}
	if err = json.Unmarshal(body, &rs); err != nil {
		return err
	}
	if rs.Status != "Ok" {
		return fmt.Errorf("public report function failed: %v", rs.Error)
	}
	return nil
}

func (pt *protoHttp) Close() error {
	pt.mu.Lock()
	pt.http = nil
	pt.mu.Unlock()
	return nil
}
