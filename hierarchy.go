package telemetry

import (
	"strings"
	"time"
)

type Hierarchy map[string]interface{}

func (h Hierarchy) Insert(name string, id FieldID) {
	a := strings.Split(name, ".")
	b := a[len(a)-1]
	for _, x := range a[:len(a)-1] {
		q, ok := h[x]
		if !ok {
			q = Hierarchy{}
			h[x] = q
		}
		h = q.(Hierarchy)
	}
	h[b] = id
}

func (h Hierarchy) Encode(name string, fields []interface{}) Packet {
	t := Packet{}
	t.Telemetry.Channel = name
	t.Telemetry.Data = h.encode(fields)
	t.Telemetry.When = time.Now().Format(time.RFC3339)
	return t
}

func (h Hierarchy) encode(fields []interface{}) map[string]interface{} {
	m := map[string]interface{}{}
	for n, e := range h {
		switch q := e.(type) {
		case FieldID:
			if g := fields[q]; g != nil {
				m[n] = g
			}
		case Hierarchy:
			m[n] = q.encode(fields)
		}
	}
	return m
}
