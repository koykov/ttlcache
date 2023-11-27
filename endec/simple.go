package endec

import (
	"bytes"
)

type simpleEnc[T any] interface {
	Encode(v T) error
}

type simpleDec[T any] interface {
	Decode(v T) error
}

type simple[T any] struct {
	buf bytes.Buffer
	enc simpleEnc[T]
	dec simpleDec[T]
}

func (t simple[T]) Encode(dst []byte, v T) ([]byte, int, error) {
	t.buf.Reset()
	err := t.enc.Encode(v)
	if err != nil {
		return dst, t.buf.Len(), err
	}
	dst = append(dst, t.buf.Bytes()...)
	return dst, t.buf.Len(), nil
}

func (t simple[T]) Decode(v *T, p []byte) error {
	t.buf.Reset()
	_, _ = t.buf.Write(p)
	err := t.dec.Decode(*v)
	return err
}
