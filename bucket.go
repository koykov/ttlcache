package ttlcache

import "sync"

type bucket[T any] struct {
	conf *Config
	id   string
	size uint64
	mux  sync.RWMutex
	idx  map[uint64]uint
	buf  []entry[T]

	null T
}

func (b *bucket[T]) set(hkey uint64, value T) error {
	now := b.clk().Now()
	b.mux.Lock()
	defer b.mux.Unlock()
	if i, ok := b.idx[hkey]; ok {
		b.buf[i] = entry[T]{
			payload:   value,
			hkey:      hkey,
			timestamp: b.conf.Clock.Now().Add(b.conf.TTLInterval).UnixNano(),
		}
		b.mw().Set(b.id, b.clk().Now().Sub(now))
		return nil
	}
	b.buf = append(b.buf, entry[T]{
		payload:   value,
		hkey:      hkey,
		timestamp: b.conf.Clock.Now().Add(b.conf.TTLInterval).UnixNano(),
	})
	b.idx[hkey] = uint(len(b.buf) - 1)
	b.mw().Set(b.id, b.clk().Now().Sub(now))
	return nil
}

func (b *bucket[T]) get(hkey uint64) (T, error) {
	now := b.clk().Now()
	b.mux.RLock()
	defer b.mux.RUnlock()
	var (
		i  uint
		ok bool
	)
	if i, ok = b.idx[hkey]; ok {
		e := &b.buf[i]
		now1 := b.clk().Now()
		if e.timestamp < now1.UnixNano() {
			b.mw().Expire(b.id)
			return b.null, ErrExpire
		}
		b.mw().Hit(b.id, now1.Sub(now))
		return e.payload, nil
	}
	b.mw().Miss(b.id)
	return b.null, ErrNotFound
}

func (b *bucket[T]) evict() error {
	now := b.clk().Now().UnixNano()
	b.mux.Lock()
	defer b.mux.Unlock()
	for i := 0; i < len(b.buf); i++ {
		if now-b.buf[i].timestamp > int64(b.conf.TTLInterval) {
			l := len(b.buf)
			old := b.buf[i].hkey
			b.buf[i] = b.buf[l-1]
			b.buf = b.buf[:l-1]
			if i < len(b.buf) {
				// Edge case: has been deleted last item.
				b.idx[b.buf[i].hkey] = uint(i)
			}
			delete(b.idx, old)
		}
	}
	return nil
}

func (b *bucket[T]) mw() MetricsWriter {
	return b.conf.MetricsWriter
}

func (b *bucket[T]) clk() Clock {
	return b.conf.Clock
}
