package toolkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spacemeshos/telemetry"
	"net/http"
)

type Reporter struct {
	telemetry.Protocol
	Verbose  func(fmt string, a ...interface{})
	Validate func(pk telemetry.Packet) error
}

func (x Reporter) Report(w http.ResponseWriter, r *http.Request) {
	e := HttpResult{"Ok", ""}
	if err := x.report(r); err != nil {
		e.Status = "Error"
		e.Error = err.Error()
	}
	json.NewEncoder(w).Encode(e)
}

func (x Reporter) report(r *http.Request) error {
	verbose := func(string, ...interface{}) {}
	if x.Verbose != nil {
		verbose = x.Verbose
	}
	pk := telemetry.Packet{}
	if err := json.NewDecoder(r.Body).Decode(&pk); err != nil {
		verbose("failed to decode packet: %v", err.Error())
		return err
	}

	if x.Verbose != nil {
		bf := &bytes.Buffer{}
		encoder := json.NewEncoder(bf)
		encoder.SetIndent("", "\t")
		_ = encoder.Encode(pk)
		verbose("INPUT packet is:\n%v", bf.String())
	}

	// Decompress if required
	if len(pk.Telemetry.Compressed) != 0 {
		verbose("compressed payload found")
		if len(pk.Telemetry.Data) != 0 {
			err := fmt.Errorf("compressed payload is not allowed in telemetry if uncompressed one exists")
			verbose("error: %v", err.Error())
			return err
		}
		d, err := Decompress(pk.Telemetry.Compressed)
		if err != nil {
			err = fmt.Errorf("decompression failed: %v", err.Error())
			verbose("error: %v", err.Error())
			return err
		}
		pk.Telemetry.Data = d
	}

	// Validate
	if x.Validate != nil {
		if err := x.Validate(pk); err != nil {
			verbose("error: %v", err.Error())
			return err
		}
	}

	pk.Telemetry.Compressed = ""
	pk.Telemetry.Tags = append(pk.Telemetry.Tags, "external")

	if x.Verbose != nil {
		bf := &bytes.Buffer{}
		encoder := json.NewEncoder(bf)
		encoder.SetIndent("", "\t")
		_ = encoder.Encode(pk)
		verbose("OUTPUT packet is:\n%v", bf.String())
	}

	// Pass to backend
	if err := x.Protocol.Send(pk); err != nil {
		err = fmt.Errorf("failed to pass the packet through: %v", err.Error())
		verbose("error: %v", err.Error())
		return err
	}

	return nil
}
