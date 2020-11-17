package telemetry

import (
	"fmt"
	"strings"
	"sync"
)

type ChannelObject struct {
	onSliceBegin, onSliceEnd []Handler

	hierarchy Hierarchy

	fields map[string]FieldID
	slice  *sliceObject

	hmu   sync.Mutex
	mu    sync.Mutex
	index uint64
	wc    WC

	protocol Protocol
	name     string
	origin   string
	onerr    func(error)
}

type FieldID uint64

type Handler func(Slice)

func New(name string, proto ProtocolFactory) *ChannelObject {
	return Channel{name, "", proto, nil}.New()
}

func ForEvery(f func(c *ChannelObject), channels ...*ChannelObject) {
	if len(channels) > 0 {
		wg := sync.WaitGroup{}
		wg.Add(len(channels))
		for _, c := range channels {
			go func(c *ChannelObject) {
				defer wg.Done()
				f(c)
			}(c)
		}
		wg.Wait()
		return
	}
}

type Channel struct {
	Name    string
	Origin  string
	Proto   ProtocolFactory
	OnError func(error)
}

type ProtocolFactory interface {
	New() Protocol
}

func (c Channel) New() *ChannelObject {
	return &ChannelObject{
		hierarchy: Hierarchy{},
		fields:    map[string]FieldID{},
		protocol:  c.Proto.New(),
		name:      c.Name,
		origin:    c.Origin,
		onerr:     c.OnError,
		slice:     &sliceObject{},
	}
}

func (c *ChannelObject) init() {
	c.slice = &sliceObject{fields: make([]interface{}, len(c.fields))}
	for _, h := range c.onSliceBegin {
		h(c.slice)
	}
}

func (c *ChannelObject) Commit() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, h := range c.onSliceEnd {
		h(c.slice)
	}
	go func(o *sliceObject, index uint64) {
		c.hmu.Lock()
		pk := c.hierarchy.Encode(c.name, o.fields)
		c.hmu.Unlock()
		c.wc.Wait(index)
		pk.Telemetry.Index = int(index)
		pk.Telemetry.Origin = c.origin
		if err := c.protocol.Send(pk); err != nil {
			if c.onerr != nil {
				c.onerr(err)
			}
		}
		c.wc.Inc()
	}(c.slice, c.index)
	c.index++
	c.init()
}

func (c *ChannelObject) Drop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.index++
	c.init()
}

func (c *ChannelObject) Wait() {
	c.wc.Wait(c.index)
}

func (c *ChannelObject) OnSliceEnd(handler Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onSliceEnd = append(c.onSliceEnd, handler)
}

func (c *ChannelObject) OnSliceBegin(handler Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onSliceBegin = append(c.onSliceBegin, handler)
	handler(c.slice)
}

func (c *ChannelObject) UpdateSlice(handler Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	handler(c.slice)
}

func (c *ChannelObject) Bind(name string) (id FieldID, ok bool) {
	s := strings.ToLower(name)
	c.mu.Lock()
	id, ok = c.fields[s]
	if !ok {
		id = FieldID(len(c.fields))
		c.fields[s] = id
		ok = true
	}
	c.mu.Unlock()
	c.hmu.Lock()
	c.hierarchy.Insert(name, id)
	c.hmu.Unlock()
	return
}

func (c *ChannelObject) Close() error {
	if c.protocol != nil {
		return c.protocol.Close()
	}
	return nil
}

func WaitAndClose(channels ...*ChannelObject) {
	ForEvery(func(c *ChannelObject) {
		c.Wait()
		c.Close()
	}, channels...)
}

func WaitFor(channels ...*ChannelObject) {
	ForEvery(func(c *ChannelObject) {
		c.Wait()
	}, channels...)
}

func Close(channels ...*ChannelObject) {
	ForEvery(func(c *ChannelObject) {
		c.Close()
	}, channels...)
}

func Commit(channels ...*ChannelObject) {
	ForEvery(func(c *ChannelObject) {
		c.Commit()
	}, channels...)
}

func Drop(channels ...*ChannelObject) {
	ForEvery(func(c *ChannelObject) {
		c.Drop()
	}, channels...)
}

type Slice interface {
	Get(id FieldID) (v interface{}, ok bool)
	Add(id FieldID, v interface{})
}

type sliceObject struct {
	fields []interface{}
}

func (s sliceObject) Get(id FieldID) (v interface{}, ok bool) {
	if len(s.fields) <= int(id) {
		return nil, false
	}
	return s.fields[id], true
}

func (s *sliceObject) Add(id FieldID, v interface{}) {
	if len(s.fields) <= int(id) {
		s.fields = make([]interface{}, id+1)
	}
	s.fields[id] = v
	return
}

type Binds []struct {
	FieldID
	*ChannelObject
}

func BindsFor(what func(id FieldID, channel *ChannelObject), name string, channels ...*ChannelObject) Binds {
	b := make(Binds, len(channels))
	for i, c := range channels {
		id, ok := c.Bind(name)
		if !ok {
			panic(fmt.Errorf("telemetry Binds %v is not unique", name))
		}
		what(id, c)
		b[i].FieldID = id
		b[i].ChannelObject = c
	}
	return b
}

func (b Binds) Update(op func(Slice, FieldID)) {
	for _, p := range b {
		p.ChannelObject.UpdateSlice(func(s Slice) {
			op(s, p.FieldID)
		})
	}
}

type Protocol interface {
	Send(pk Packet) error
	Close() error
}

type Packet struct {
	Telemetry struct {
		Channel    string                 `json:"channel"`
		Origin     string                 `json:"origin"`
		Index      int                    `json:"index"`
		When       string                 `json:"when"`
		Tags       []string               `json:"tags,omitempty"`
		Data       map[string]interface{} `json:"data,omitempty"`
		Compressed string                 `json:"compressed,omitempty"`
	} `json:"telemetry"`
}
