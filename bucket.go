package ttlcache

import "sync"

type bucket[T any] struct {
	conf *Config
	id   string
	size uint64
	mux  sync.RWMutex
	idx  map[uint64]uint
	buf  []entry[T]
}

func (b *bucket[T]) set(hkey uint64, value T) error {
	b.mux.Lock()
	defer b.mux.Unlock()
	now := b.clk().Now()
	if i, ok := b.idx[hkey]; ok {
		b.buf[i] = entry[T]{
			payload:   value,
			hkey:      hkey,
			timestamp: b.conf.Clock.Now().UnixNano(),
		}
		b.mw().Set(b.id, b.clk().Now().Sub(now))
		return nil
	}
	b.buf = append(b.buf, entry[T]{
		payload:   value,
		hkey:      hkey,
		timestamp: b.conf.Clock.Now().UnixNano(),
	})
	b.idx[hkey] = uint(len(b.buf))
	b.mw().Set(b.id, b.clk().Now().Sub(now))
	return nil
}

func (b *bucket[T]) get(hkey uint64) (T, error) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	var (
		i  uint
		ok bool
	)
	now := b.clk().Now()
	if i, ok = b.idx[hkey]; ok {
		e := &b.buf[i]
		now1 := b.clk().Now()
		if e.timestamp < now1.UnixNano() {
			b.mw().Expire(b.id)
			return any(nil), ErrExpire
		}
		b.mw().Hit(b.id, now1.Sub(now))
		return e.payload, nil
	}
	b.mw().Miss(b.id)
	return any(nil), ErrNotFound
}

func (b *bucket[T]) evict() error {
	// ...
	return nil
}

func (b *bucket[T]) mw() MetricsWriter {
	return b.conf.MetricsWriter
}

func (b *bucket[T]) clk() Clock {
	return b.conf.Clock
}
