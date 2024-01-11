package endec

import (
	"github.com/koykov/bytealg"
	"github.com/koykov/ttlcache"
)

type Protobuf[T any] struct{}

func (t Protobuf[T]) Encode(dst []byte, v T) ([]byte, int, error) {
	var m ttlcache.MarshallerTo
	switch x := any(v).(type) {
	case ttlcache.MarshallerTo:
		m = x
	default:
		return dst, 0, ErrPBUnsupportedType
	}
	dst = bytealg.Grow(dst, m.Size())
	n, err := m.MarshalTo(dst)
	return dst, n, err
}

func (t Protobuf[T]) Decode(v *T, p []byte) error {
	var m ttlcache.Unmarshaller
	switch x := any(*v).(type) {
	case ttlcache.Unmarshaller:
		m = x
	default:
		return ErrPBUnsupportedType
	}
	err := m.Unmarshal(p)
	return err
}
