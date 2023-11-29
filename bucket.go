package ttlcache

import (
	"io"
	"strconv"
	"sync"
)

const dumpBufSize = 4096

type bucket[T any] struct {
	conf *Config[T]
	id   string
	size uint64
	mux  sync.RWMutex
	idx  map[uint64]uint
	buf  []entry[T]
	bbuf []byte

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

func (b *bucket[T]) delete(hkey uint64) error {
	b.mux.Lock()
	defer b.mux.Unlock()
	if idx, ok := b.idx[hkey]; ok {
		b.evictLF(idx)
	}
	return nil
}

func (b *bucket[T]) extract(hkey uint64) (T, error) {
	now := b.clk().Now()
	b.mux.Lock()
	defer b.mux.Unlock()
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
		b.evictLF(i)
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
			b.evictLF(uint(i))
		}
	}
	return nil
}

func (b *bucket[T]) evictLF(idx uint) {
	l := len(b.buf)
	old := b.buf[idx].hkey
	b.buf[idx] = b.buf[l-1]
	b.buf = b.buf[:l-1]
	if idx < uint(len(b.buf)) {
		// Edge case: has been deleted last item.
		b.idx[b.buf[idx].hkey] = idx
	}
	delete(b.idx, old)
}

func (b *bucket[T]) dump(w io.Writer) (err error) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	b.bbuf = b.bbuf[:0]
	for i := 0; i < len(b.buf); i++ {
		e := &b.buf[i]
		b.bbuf = strconv.AppendUint(b.bbuf, e.hkey, 10)
		b.bbuf = strconv.AppendInt(b.bbuf, e.timestamp, 10)
		if b.bbuf, _, err = b.conf.DumpEncoder.Encode(b.bbuf, e.payload); err != nil {
			return
		}
		if len(b.bbuf) > dumpBufSize {
			if _, err = w.Write(b.bbuf); err != nil {
				return
			}
			b.bbuf = b.bbuf[:0]
		}
	}
	if len(b.bbuf) > 0 {
		_, err = w.Write(b.bbuf)
	}
	return
}

func (b *bucket[T]) reset() error {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.buf = b.buf[:0]
	for k := range b.idx {
		delete(b.idx, k)
	}
	return nil
}

func (b *bucket[T]) close() error {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.buf, b.idx = nil, nil
	return nil
}

func (b *bucket[T]) mw() MetricsWriter {
	return b.conf.MetricsWriter
}

func (b *bucket[T]) clk() Clock {
	return b.conf.Clock
}
