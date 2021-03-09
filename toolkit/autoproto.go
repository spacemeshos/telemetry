package toolkit

import (
	"fmt"
	"github.com/spacemeshos/telemetry"
	"strings"
	"time"
)

type Auto struct {
	Endpoint   string
	Timeout    time.Duration
	Compressed bool
}

func (c Auto) New() telemetry.Protocol {
	if strings.HasPrefix(c.Endpoint, "http://") {
		return Http{Endpoint: c.Endpoint, Timeout: c.Timeout, Compressed: c.Compressed}.New()
	} else if strings.HasPrefix(c.Endpoint, "tcp://") {
		return Tcp{Endpoint: c.Endpoint[6:], Timeout: c.Timeout}.New()
	} else {
		panic(fmt.Errorf("unknown telemetry protocol for %v", c.Endpoint))
	}
}
