package ttlcache

import "sync"

type bucket[T any] struct {
	conf *Config
	size uint64
	mux  sync.RWMutex
	idx  map[uint64]uint
	buf  []entry[T]
}

func (b *bucket[T]) set(hkey uint64, value T) error {
	b.mux.Lock()
	defer b.mux.Unlock()
	if i, ok := b.idx[hkey]; ok {
		b.buf[i] = entry[T]{
			payload:   value,
			hkey:      hkey,
			timestamp: b.conf.Clock.Now().UnixNano(),
		}
		return nil
	}
	b.buf = append(b.buf, entry[T]{
		payload:   value,
		hkey:      hkey,
		timestamp: b.conf.Clock.Now().UnixNano(),
	})
	b.idx[hkey] = uint(len(b.buf))
	return nil
}

func (b *bucket[T]) get(hkey uint64) (T, error) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	var (
		i  uint
		ok bool
	)
	if i, ok = b.idx[hkey]; ok {
		return b.buf[i].payload, nil
	}
	return any(nil), ErrNotFound
}
