package ttlcache

import "sync"

type Cache[T any] struct {
	o       sync.Once
	conf    *Config
	buckets []bucket[T]

	err error
}

func (c *Cache[T]) init() {
	//
}

func (c *Cache[T]) Set(key string, value T) error {
	c.o.Do(c.init)
	if c.err != nil {
		return c.err
	}
	hkey := c.conf.Hasher.Sum64(key)
	b := &c.buckets[hkey%uint64(c.conf.Buckets)]
	return b.set(hkey, value)
}

func (c *Cache[T]) Get(key string) (T, error) {
	c.o.Do(c.init)
	if c.err != nil {
		return any(nil), c.err
	}
	hkey := c.conf.Hasher.Sum64(key)
	b := &c.buckets[hkey%uint64(c.conf.Buckets)]
	return b.get(hkey)
}
