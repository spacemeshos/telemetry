package telemetry

import (
	"encoding/json"
	"math"
)

type Counter struct{ Binds }

func BindCounter(name string, channels ...*ChannelObject) Counter {
	return Counter{BindsFor(func(id FieldID, channel *ChannelObject) {
		channel.OnSliceBegin(func(slice Slice) {
			slice.Add(id, &CounterValue{})
		})
	}, name, channels...)}
}

func BindCounterIf(exp bool, name string) Counter {
	if exp {
		return BindCounter(name)
	}
	return Counter{}
}

func (c Counter) Inc(val uint) {
	c.Update(func(s Slice, id FieldID) {
		if v, ok := s.Get(id); ok {
			v.(*CounterValue).Inc(val)
		}
	})
}

type CounterValue struct{ value uint64 }

func (c *CounterValue) Inc(val uint) {
	k := c.value + uint64(val)
	if k < c.value { // saturate on overflow
		k = math.MaxUint64
	}
	c.value = k
}

func (c *CounterValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}
