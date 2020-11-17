package telemetry

import (
	"encoding/json"
)

type Number struct{ Binds }
type FuncNum func()float64
type DfltNum float64

func (f FuncNum) Get() float64 { if f != nil { return f() }; return 0 }
func (f DfltNum) Get() float64 { return float64(f) }

func BindNumber(name string, initializer FuncNum, channels ...*ChannelObject) Number {
	return Number{BindsFor(func(id FieldID, channel *ChannelObject) {
		channel.OnSliceBegin(func(slice Slice) {
			slice.Add(id, &NumberValue{initializer.Get() } )
		})
	}, name, channels...)}
}

func BindNumberIf(exp bool, name string, initializer FuncNum, channels ...*ChannelObject) Number {
	if exp {
		return BindNumber(name, initializer, channels...)
	}
	return Number{}
}

func BindAutoNumber(name string, getter FuncNum, channels ...*ChannelObject) bool {
	BindsFor(func(id FieldID, channel *ChannelObject) {
		channel.OnSliceEnd(func(slice Slice) {
			slice.Add(id, &NumberValue{getter.Get()})
		})
	}, name, channels...)
	return true
}

func BindAutoNumberIf(exp bool, name string, getter func() float64, channels ...*ChannelObject) bool {
	if exp {
		return BindAutoNumber(name, getter, channels...)
	}
	return true
}

func ConstantNumber(name string, value float64, channels ...*ChannelObject) bool {
	return BindAutoNumber(name, DfltNum(value).Get, channels...)
}

func (c Number) Set(val float64) {
	c.Update(func(s Slice, id FieldID) {
		if v, ok := s.Get(id); ok {
			v.(*NumberValue).Set(val)
		}
	})
}

type NumberValue struct {
	value float64
}

func (c *NumberValue) Set(val float64) {
	c.value = val
}

func (c *NumberValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}
