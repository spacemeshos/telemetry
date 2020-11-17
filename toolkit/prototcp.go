package toolkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spacemeshos/telemetry"
	"net"
	"sync"
	"time"
)

type protoTcp struct {
	mu     sync.Mutex
	conn   net.Conn
	closed bool

	endpoint string
	timeout  time.Duration
}

type Tcp struct {
	Endpoint string
	Timeout  time.Duration
}

func (t Tcp) New() telemetry.Protocol {
	return &protoTcp{
		endpoint: t.Endpoint,
		timeout:  t.Timeout,
	}
}

func (pt *protoTcp) Send(pk telemetry.Packet) error {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	if pt.closed {
		return fmt.Errorf("tcp protocol is closed")
	}
	if pt.conn == nil {
		conn, err := net.DialTimeout("tcp", pt.endpoint, pt.timeout)
		if err != nil {
			return err
		}
		pt.conn = conn
	}
	bf := &bytes.Buffer{}
	err := json.NewEncoder(bf).Encode(pk)
	if err != nil {
		pt.close()
		return err
	}
	if pt.timeout != 0 {
		err = pt.conn.SetWriteDeadline(time.Now().Add(pt.timeout))
		if err != nil {
			pt.close()
			return err
		}
	}
	_, err = pt.conn.Write(bf.Bytes())
	if err != nil {
		pt.close()
		return err
	}
	return nil
}

func (pt *protoTcp) Close() error {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.closed = true
	return pt.close()
}

func (pt *protoTcp) close() (err error) {
	if pt.conn != nil {
		err = pt.conn.Close()
		pt.conn = nil
	}
	return
}
