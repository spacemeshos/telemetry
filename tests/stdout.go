package tests

import (
	"encoding/json"
	"github.com/spacemeshos/telemetry"
	"os"
)

type StdOut struct {
}

func (StdOut) Send(pk telemetry.Packet) error {
	js, err := json.Marshal(pk)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(append(js, '\n'))
	return err
}
