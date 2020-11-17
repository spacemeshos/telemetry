package telemetry

import (
	"encoding/json"
)

type Statistic struct{ Binds }

func BindStatistic(name string, channels ...*ChannelObject) Statistic {
	return Statistic{BindsFor(func(id FieldID, channel *ChannelObject) {
		channel.OnSliceBegin(func(slice Slice) {
			slice.Add(id, &StatisticValue{})
		})
	}, name, channels...)}
}

func BindStatisticIf(exp bool, name string, channels ...*ChannelObject) Statistic {
	if exp {
		return BindStatistic(name, channels...)
	}
	return Statistic{}
}

func (c Statistic) Set(val float64) {
	c.Update(func(s Slice, id FieldID) {
		if v, ok := s.Get(id); ok {
			v.(*StatisticValue).Set(val)
		}
	})
}

func (c Statistic) Inc(val uint) {
	c.Update(func(s Slice, id FieldID) {
		if v, ok := s.Get(id); ok {
			v.(*StatisticValue).Inc(val)
		}
	})
}

type StatisticValue struct {
	value                  float64
	min, max, total, count float64
}

func (c *StatisticValue) Set(val float64) {
	c.value = val
	if c.count > 0 {
		if val < c.min {
			c.min = val
		}
		if val > c.max {
			c.max = val
		}
	} else {
		c.min = val
		c.max = val
	}
	c.count++
	c.total += val
}

func (c *StatisticValue) Inc(val uint) {
	c.Set(c.value + float64(val))
}

func (c *StatisticValue) MarshalJSON() ([]byte, error) {
	v := struct {
		Average float64 `json:"average"`
		Min     float64 `json:"min"`
		Max     float64 `json:"max"`
		Count   float64 `json:"count"`
	}{c.total / c.count, c.min, c.max, c.count}
	return json.Marshal(v)
}
