package ttlcache

import "sync"

type bucket[T any] struct {
	size uint64
	mux  sync.RWMutex
	idx  map[uint64]uint
	buf  []entry[T]
}

func (b *bucket[T]) set(hkey uint64, value T) error {
	_, _ = hkey, value
	return nil
}

func (b *bucket[T]) get(hkey uint64) (T, error) {
	_ = hkey
	return any(nil), nil
}
