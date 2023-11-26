package endec

import (
	"bytes"
	"encoding/gob"
)

type GOB[T any] struct {
	buf bytes.Buffer
	enc *gob.Encoder
	dec *gob.Decoder
}

func (t GOB[T]) Encode(dst []byte, v T) ([]byte, int, error) {
	t.buf.Reset()
	if t.enc == nil {
		t.enc = gob.NewEncoder(&t.buf)
	}
	err := t.enc.Encode(v)
	if err != nil {
		return dst, t.buf.Len(), err
	}
	dst = append(dst, t.buf.Bytes()...)
	return dst, t.buf.Len(), nil
}

func (t GOB[T]) Decode(v *T, p []byte) error {
	t.buf.Reset()
	if t.dec == nil {
		t.dec = gob.NewDecoder(&t.buf)
	}
	_, _ = t.buf.Write(p)
	err := t.dec.Decode(*v)
	return err
}
