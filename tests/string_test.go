package tests

import (
	"github.com/spacemeshos/telemetry"
	"gotest.tools/assert"
	"testing"
)

func Test_StringSet(t *testing.T) {
	f := "string"
	to := &TestTelemetryOut{
		Channel: "String Set Test",
		Data: []J{
			{"string": "hello"},
			{"string": "world"},
			{"string": "here"},
		}}
	c := telemetry.New(to.Channel, to)
	v := telemetry.BindString(f, nil, c)
	v.Set(to.Data[0][f].(string))
	c.Commit()
	v.Set(to.Data[1][f].(string))
	c.Commit()
	v.Set("some shit here")
	v.Set(to.Data[2][f].(string))
	c.Commit()
	telemetry.WaitAndClose(c)
	assert.Assert(t, to.Ok(), to.What())
}

func Test_StringDefault(t *testing.T) {
	f := "string"
	to := &TestTelemetryOut{
		Channel: "String Set Test",
		Data: []J{
			{"string": "default"},
			{"string": "new value"},
		}}
	c := telemetry.New(to.Channel, to)
	v := telemetry.BindString(f, telemetry.DfltStr(to.Data[0][f].(string)).Get, c)
	c.Commit()
	v.Set(to.Data[1][f].(string))
	c.Commit()
	telemetry.WaitAndClose(c)
	assert.Assert(t, to.Ok(), to.What())
}

func Test_StringConstant(t *testing.T) {
	f := "string"
	cf := "constnat"
	to1 := &TestTelemetryOut{
		Channel: "Test Channel 1",
		Data: []J{
			{"string": cf},
			{"string": cf},
		}}
	to2 := &TestTelemetryOut{
		Channel: "Test Channel 2",
		Data: []J{
			{"string": cf},
		}}

	c1 := telemetry.New(to1.Channel, to1)
	c2 := telemetry.New(to2.Channel, to2)
	telemetry.ConstantString(f, cf, c1, c2)
	telemetry.Commit(c1)
	telemetry.Commit(c1, c2)
	telemetry.WaitAndClose(c1, c2)
	assert.Assert(t, to1.Ok(), to1.What())
	assert.Assert(t, to2.Ok(), to2.What())
}
