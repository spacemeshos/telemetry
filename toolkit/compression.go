package toolkit

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
)

func Decompress(data string) (map[string]interface{}, error) {
	r, err := gzip.NewReader(base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(data)))
	if err != nil {
		return nil, err
	}
	m := map[string]interface{}{}
	defer r.Close()
	if err = json.NewDecoder(r).Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}

func Compress(data map[string]interface{}) (string, error) {
	bf := bytes.Buffer{}
	b64 := base64.NewEncoder(base64.StdEncoding, &bf)
	w := gzip.NewWriter(b64)
	e := json.NewEncoder(w)
	if err := e.Encode(data); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	if err := b64.Close(); err != nil {
		return "", err
	}
	return bf.String(), nil
}
