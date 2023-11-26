package endec

import (
	"bytes"
	"encoding/json"
)

type JSON[T any] struct {
	buf bytes.Buffer
	enc *json.Encoder
	dec *json.Decoder
}

func (t JSON[T]) Encode(dst []byte, v T) ([]byte, int, error) {
	t.buf.Reset()
	if t.enc == nil {
		t.enc = json.NewEncoder(&t.buf)
	}
	err := t.enc.Encode(v)
	if err != nil {
		return dst, t.buf.Len(), err
	}
	dst = append(dst, t.buf.Bytes()...)
	return dst, t.buf.Len(), nil
}

func (t JSON[T]) Decode(v *T, p []byte) error {
	t.buf.Reset()
	if t.dec == nil {
		t.dec = json.NewDecoder(&t.buf)
	}
	_, _ = t.buf.Write(p)
	err := t.dec.Decode(*v)
	return err
}
