package ttlcache

import "sync"

type Cache[T any] struct {
	o       sync.Once
	conf    *Config
	buckets []bucket[T]
}
