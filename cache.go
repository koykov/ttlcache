package ttlcache

import (
	"strconv"
	"sync"
	"time"
)

type Cache[T any] struct {
	o       sync.Once
	conf    *Config
	buckets []bucket[T]

	err error
}

func New[T any](conf *Config) (*Cache[T], error) {
	c := &Cache[T]{
		conf: conf.Copy(),
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

	var bsize uint64
	if c.conf.Size > 0 {
		bsize = c.conf.Size / uint64(c.conf.Buckets)
	}
	c.buckets = make([]bucket[T], 0, c.conf.Buckets)
	for i := uint(0); i < c.conf.Buckets; i++ {
		c.buckets = append(c.buckets, bucket[T]{
			conf: c.conf,
			id:   strconv.Itoa(int(i)),
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

func (c *Cache[T]) bulkEvict() error {
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
					_ = c.buckets[idx].evict()
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

func (c *Cache[T]) l() Logger {
	return c.conf.Logger
}
