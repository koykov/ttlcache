package ttlcache

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Cache[T any] struct {
	o       sync.Once
	status  uint32
	conf    *Config
	buckets []bucket[T]
	null    T

	err error
}

func New[T any](conf *Config) (*Cache[T], error) {
	c := &Cache[T]{
		status: cacheStatusActive,
		conf:   conf.Copy(),
	}
	c.o.Do(c.init)
	if c.err != nil {
		return nil, c.err
	}
	return c, nil
}

func (c *Cache[T]) init() {
	if c.conf == nil {
		c.err = ErrNoConfig
		return
	}
	if c.conf.Hasher == nil {
		c.err = ErrNoHasher
		return
	}
	if c.conf.Buckets == 0 {
		c.err = ErrNoBuckets
		return
	}

	if c.conf.MetricsWriter == nil {
		c.conf.MetricsWriter = DummyMetrics{}
	}

	if c.conf.Clock == nil {
		c.conf.Clock = &NativeClock{}
	}
	if !c.conf.Clock.Active() {
		c.conf.Clock.Start()
	}

	var bsize uint64
	if c.conf.Size > 0 {
		bsize = c.conf.Size / uint64(c.conf.Buckets)
	}
	c.buckets = make([]bucket[T], 0, c.conf.Buckets)
	for i := uint(0); i < c.conf.Buckets; i++ {
		c.buckets = append(c.buckets, bucket[T]{
			conf: c.conf,
			id:   strconv.Itoa(int(i)),
			idx:  make(map[uint64]uint, bsize),
			buf:  make([]entry[T], 0, bsize),
			size: bsize,
		})
	}

	if c.conf.TTLInterval > 0 {
		if c.conf.TTLInterval < time.Second {
			c.err = ErrShortTTL
			return
		}
		if c.conf.EvictInterval == 0 {
			c.conf.EvictInterval = c.conf.TTLInterval / 2
		}
		if c.conf.EvictWorkers == 0 {
			c.conf.EvictWorkers = defaultEvictWorkers
		}
		c.conf.Clock.Schedule(c.conf.EvictInterval, func() {
			if err := c.bulkEvict(); err != nil && c.l() != nil {
				c.l().Printf("eviction failed with error %s\n", err.Error())
			}
		})
	}
}

func (c *Cache[T]) Set(key string, value T) error {
	c.o.Do(c.init)
	if c.err != nil {
		return c.err
	}
	if err := c.checkCache(cacheStatusActive); err != nil {
		return err
	}
	hkey := c.conf.Hasher.Sum64(key)
	b := &c.buckets[hkey%uint64(c.conf.Buckets)]
	return b.set(hkey, value)
}

func (c *Cache[T]) Get(key string) (T, error) {
	c.o.Do(c.init)
	if c.err != nil {
		return c.null, c.err
	}
	if err := c.checkCache(cacheStatusActive); err != nil {
		return c.null, err
	}
	hkey := c.conf.Hasher.Sum64(key)
	b := &c.buckets[hkey%uint64(c.conf.Buckets)]
	return b.get(hkey)
}

func (c *Cache[T]) Delete(key string) error {
	c.o.Do(c.init)
	if c.err != nil {
		return c.err
	}
	if err := c.checkCache(cacheStatusActive); err != nil {
		return err
	}
	hkey := c.conf.Hasher.Sum64(key)
	b := &c.buckets[hkey%uint64(c.conf.Buckets)]
	return b.delete(hkey)
}

func (c *Cache[T]) Extract(key string) (T, error) {
	c.o.Do(c.init)
	if c.err != nil {
		return c.null, c.err
	}
	if err := c.checkCache(cacheStatusActive); err != nil {
		return c.null, err
	}
	hkey := c.conf.Hasher.Sum64(key)
	b := &c.buckets[hkey%uint64(c.conf.Buckets)]
	return b.extract(hkey)
}

func (c *Cache[T]) Close() error {
	atomic.StoreUint32(&c.status, cacheStatusClosed)
	c.conf.Clock.Stop()
	return c.bulkClose()
}

func (c *Cache[T]) Reset() error {
	return c.bulkReset()
}

func (c *Cache[T]) bulkEvict() error {
	return c.bulkExec(func(b *bucket[T]) error {
		return b.evict()
	})
}

func (c *Cache[T]) bulkClose() error {
	return c.bulkExec(func(b *bucket[T]) error {
		return b.close()
	})
}

func (c *Cache[T]) bulkReset() error {
	return c.bulkExec(func(b *bucket[T]) error {
		return b.reset()
	})
}

func (c *Cache[T]) bulkExec(fn func(b *bucket[T]) error) error {
	var workers = c.conf.EvictWorkers
	if c.conf.Buckets < workers {
		workers = c.conf.Buckets
	}
	bucketQueue := make(chan uint, workers)
	var wg sync.WaitGroup
	for i := uint(0); i < workers; i++ {
		wg.Add(1)
		go func(i uint) {
			defer wg.Done()
			for {
				if idx, ok := <-bucketQueue; ok {
					_ = fn(&c.buckets[idx])
					continue
				}
				break
			}
		}(i)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := uint(0); i < workers; i++ {
			bucketQueue <- i
		}
		close(bucketQueue)
	}()

	wg.Wait()

	return nil
}

func (c *Cache[T]) checkCache(allow uint32) error {
	if status := atomic.LoadUint32(&c.status); status&allow == 0 {
		if status == cacheStatusNil {
			return ErrBadCache
		}
		if status == cacheStatusClosed {
			return ErrCacheClosed
		}
	}
	return nil
}

func (c *Cache[T]) l() Logger {
	return c.conf.Logger
}
