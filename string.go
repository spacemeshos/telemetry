package telemetry

import (
	"encoding/json"
)

type String struct{ Binds }
type FuncStr func()string
type DfltStr string

func (f FuncStr) Get() string { if f != nil { return f() }; return "" }
func (s DfltStr) Get() string { return string(s) }

func BindString(name string, initializer FuncStr, channels ...*ChannelObject) String {
	return String{BindsFor(func(id FieldID, channel *ChannelObject) {
		channel.OnSliceBegin(func(slice Slice) {
			slice.Add(id, &StringValue{initializer.Get() })
		})
	}, name, channels...)}
}

func BindStringIf(exp bool, name string, initilizer FuncStr, channels ...*ChannelObject) String {
	if exp {
		return BindString(name, initilizer, channels...)
	}
	return String{}
}

func BindAutoString(name string, getter FuncStr, channels ...*ChannelObject) bool {
	BindsFor(func(id FieldID, channel *ChannelObject) {
		channel.OnSliceEnd(func(slice Slice) {
			slice.Add(id, &StringValue{getter.Get()})
		})
	}, name, channels...)
	return true
}

func BindAutoStringIf(exp bool, name string, getter FuncStr, channels ...*ChannelObject) bool {
	if exp {
		return BindAutoString(name, getter, channels...)
	}
	return true
}

func ConstantString(name string, value string, channels ...*ChannelObject) bool {
	return BindAutoString(name, DfltStr(value).Get, channels...)
}

func (c String) Set(val string) {
	c.Update(func(s Slice, id FieldID) {
		if v, ok := s.Get(id); ok {
			v.(*StringValue).Set(val)
		}
	})
}

type StringValue struct {
	value string
}

func (c *StringValue) Set(val string) {
	c.value = val
}

func (c *StringValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}
