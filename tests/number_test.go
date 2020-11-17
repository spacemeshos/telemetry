package tests

import (
	"github.com/spacemeshos/telemetry"
	"gotest.tools/assert"
	"testing"
)

func Test_NumberSet(t *testing.T) {
	f := "number"
	to := &TestTelemetryOut{
		Channel: "Number Set Test",
		Data: []J{
			{f: 1},
			{f: 100},
			{f: 10},
		}}
	c := telemetry.New(to.Channel, to)
	v := telemetry.BindNumber(f, nil, c)
	v.Set(1)
	c.Commit()
	v.Set(100)
	c.Commit()
	v.Set(100)
	v.Set(10)
	c.Commit()
	telemetry.WaitAndClose(c)
	assert.Assert(t, to.Ok(), to.What())
}

func Test_NumberDefault(t *testing.T) {
	f := "number"
	to := &TestTelemetryOut{
		Channel: "Number Set Test",
		Data: []J{
			{f: 10},
			{f: 100},
		}}
	c := telemetry.New(to.Channel, to)
	v := telemetry.BindNumber(f, telemetry.DfltNum(10).Get, c)
	c.Commit()
	v.Set(100)
	c.Commit()
	telemetry.WaitAndClose(c)
	assert.Assert(t, to.Ok(), to.What())
}

func Test_NumberConstant(t *testing.T) {
	f := "number"
	to1 := &TestTelemetryOut{
		Channel: "Number Set Test 1",
		Data: []J{
			{f: 50.3},
			{f: 50.3},
		}}
	to2 := &TestTelemetryOut{
		Channel: "Number Set Test 2",
		Data: []J{
			{f: 50.3},
		}}
	c1 := telemetry.New(to1.Channel, to1)
	c2 := telemetry.New(to2.Channel, to2)
	telemetry.ConstantNumber(f, 50.3, c1, c2)
	telemetry.Commit(c1)
	assert.Assert(t, to1.Ok(), to1.What())
	telemetry.Commit(c1, c2)
	telemetry.WaitAndClose(c1, c2)
	assert.Assert(t, to2.Ok(), to2.What())
}
