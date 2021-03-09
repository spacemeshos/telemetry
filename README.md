# Spacemesh Telemetry Framework

### Motivation

We needs to collect some statistics from clients during fixed time slices like layer, epoch, or/and daily. Also, we need to collect some information when new client is discovered on the network.

### Description

The telemetry framework operates with four main metaphors: channel, slice, binding, and protocol. 

* [**Channel**](channel.go#L9) is a telemetry collection target having name and set of bindings. It's can be created, reseted and committed.
* [**Slice**](channel.go#L193) is a set of binding's data holders collecting some statistics for the time period. It's not be accessed directly and exists in shadow.
* [**Biding**](channel.go#L217) is a typed object to collect required statistics used by client code. It's bounded to one or many channels and provided access to current channel's time slice.
* [**Protocol**](channel.go#247) is a network connection implementation allowing to deliver collected statistics to the telemetry server

The telemetry framework is fully extendable. It can be extended with new types of Bindings and with new protocols.

Currently, the telemetry framework has following set of bindings:

* [**Counter**](counter.go#L8) collects a numeric value, it can be incremented only. _([BindCounter](counter.go#L8))_
* [**Statistic**](statistic.go#L7) collects min, max, avg from the all value updates for during the time slice. _([BindStatistic](statistic.go#L9))_ 
* [**Number**](number.go#L14) collects the last value for the time slice. _([BindNumber](number.go#L14))_
* [**AutoNumber**](number.go#L29) collects a numeric result of a function call on slice commit.
* [**ConstantNumber**](number.go#L45) collects a numeric constant.
* [**String**](string.go#L7) collects the last string value for the time slice. _([BindString](string.go#L14))_
* [**AutoString**](string.go#L29) collects a string result of a function call on slice commit.
* [**ConstantString**](string.go#L49) collects a string constant.

Also, there are two protocols for delivering collected telemetry:

* [**toolkit.Tcp**](toolkit/prototcp.go#L22) is used to publish collected telemetry directly to logstash 
* [**toolkit.Http**](toolkit/protohttp.go#L21) is used to publish telemetry by Http-Json proxy implemented with [toolkit.Reporter](toolkit/server.go#L17)

### How to use

To use the telemetry framework you need to create at least one channel and any number of bindings. Then update values of bindings and commit channel to telemetry collecting server.

```go
c := telemetry.Channel{
    Name: "The Channel", 
    Origin: "The Client",
    Proto: toolkit.Auto{
        Endpoint: "http://telemtry.endpoint.io:8080", 
        Timeout: 3 * time.Second, 
        Compressed: True
    },
    OnError: func(e error) { log.Error(e.Error()) }
}.New()

telemetry.ConstantString("client.ID", "TheUniqueClientID", c)
nonce := telemetry.BindCounter("client.Nonce", c)
c.Commit()
nonce.Inc(1)
c.Commit()
```

### How to extend with new bindings

For example let's create new binding collecting how many times provided value was less then threshold during the time slice.

The first, we need to create new two types:

```go
type Threshold struct { Binds }
type ThresholdValue struct { count, threshold int }

func (t Threshold) Update(value int) {
    c.Update(func(s Slice, id FieldID) {
        if v, ok := s.Get(id); ok {
            th := v.(*ThresholdValue)
            if th.threshold > value {
                th.count += 1	
            }           
        }
    })
}
```

Then we need to crate binding function
```go
func BindThreshold(name string, threshold int, channels ...*ChannelObject) Threshold {
	return Threshold{BindsFor(func(id FieldID, channel *ChannelObject) {
		channel.OnSliceBegin(func(slice Slice) {
			slice.Add(id, &ThresholdValue{threshold: threshold})
		})}, name, channels...)}
}
```

Since collected threshold statistic must be converted to JSON on a commit, we need also to define how to do it

```go
func (c *ThresholdValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.threshold)
}
```

### How it works

As you can see there are two specific methods for the Channel type: [OnSliceBegin](channel.go#L126) and [OnSliceEnd](channel.go#L120). 
The first one sets up the function calling on the slice creation, and the second one sets up function calling when slice is committed just before encode all collected telemetry. 

Since a slice is no more than set of collected telemetry, you can add new data holder any time before slice encoded into telemetry packet. 
However, there are only two reasonable moments to do it: when the time slice started or ended. The first one is more usable for updatable values like a counter, and the second one is for constant or auto-collecting values.

All methods changing the channel state are thread-safe and can be called in any moment.

The application does not work directly with a slice, but with typed telemetry bindings. 
The telemetry binding must update its holder in slices (there are can be more than one slice) with [Bind](channel.go#L222)'s method [Update](channel.go#L239). 
So it's make application code independent of channels management. The specific application code must know only how to report telemetry, but not where and how often it's collected.

