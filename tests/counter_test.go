package tests

import (
	"github.com/spacemeshos/telemetry"
	"gotest.tools/assert"
	"testing"
)

func Test_Counter(t *testing.T) {
	f := "counter"
	to := &TestTelemetryOut{
		Channel: "Counter Test",
		Data: []J{
			{"counter": 1},
			{"counter": 11},
		}}
	c := telemetry.New(to.Channel, to)
	cnt := telemetry.BindCounter(f, c)
	cnt.Inc(1)
	c.Commit()
	cnt.Inc(10)
	cnt.Inc(1)
	c.Commit()
	telemetry.WaitAndClose(c)
	assert.Assert(t, to.Ok(), to.What())
}
