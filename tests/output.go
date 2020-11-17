package tests

import (
	"encoding/json"
	"fmt"
	"github.com/spacemeshos/telemetry"
	"strings"
	"sync"
)

type J = map[string]interface{}

type TestTelemetryOut struct {
	Channel string
	Data    []J
	Errors  []string
	Sender  func(js []byte) error
	mu      sync.Mutex
	wc      telemetry.WC
}

func (to *TestTelemetryOut) New() telemetry.Protocol {
	return to
}

func (to *TestTelemetryOut) report(counter int, text string, err error) error {
	e := fmt.Errorf("%d:%s:%s", counter, text, err.Error())
	to.mu.Lock()
	to.Errors = append(to.Errors, e.Error())
	to.mu.Unlock()
	return e
}

func (to *TestTelemetryOut) Ok() bool {
	return len(to.Errors) == 0
}

func (to *TestTelemetryOut) What() string {
	return strings.Join(to.Errors, "\n")
}

func (to *TestTelemetryOut) Send(pk telemetry.Packet) error {
	js, err := json.Marshal(pk)
	if err != nil {
		return to.report(pk.Telemetry.Index, "output", err)
	}
	if to.Sender != nil {
		return to.Sender(js)
	}
	return to.SenderImpl(js)
}

func (to *TestTelemetryOut) Close() error {
	return nil
}

func (to *TestTelemetryOut) SenderImpl(js []byte) error {
	pk := telemetry.Packet{}

	if err := json.Unmarshal(js, &pk); err != nil {
		err = to.report(-1, "unmarshal packet", err)
		to.wc.Inc()
		return err
	}

	var flat func(string, map[string]interface{}, map[string]string)
	flat = func(pfx string, in map[string]interface{}, out map[string]string) {
		for k, v := range in {
			switch q := v.(type) {
			case []interface{}:
				out[pfx+k] = fmt.Sprint(q)
			case map[string]interface{}:
				flat(pfx+k+".", q, out)
			default:
				out[pfx+k] = fmt.Sprint(q)
			}
		}
	}

	i := pk.Telemetry.Index
	o := map[string]string{}
	e := map[string]string{}
	flat("", pk.Telemetry.Data, o)
	flat("", to.Data[i], e)

	for k, v := range e {
		v2, ok := o[k]
		if !ok {
			to.report(i, k, fmt.Errorf("is not present in output, must be `%x`", v))
		}
		if v2 != v {
			to.report(i, k, fmt.Errorf("output `%s` is not the same as expected `%s`", v2, v))
		}
	}

	for k := range o {
		if _, ok := e[k]; !ok {
			to.report(i, k, fmt.Errorf("must not be in output"))
		}
	}

	to.wc.Inc()
	return nil
}

func (to *TestTelemetryOut) Wait() {
	to.wc.Wait(uint64(len(to.Data)))
}
