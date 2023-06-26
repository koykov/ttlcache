package ttlcache

import "sync"

type bucket[T any] struct {
	mux sync.RWMutex
	idx map[uint64]uint
	buf []entry[T]
}
