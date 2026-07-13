package ttlcache

import (
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Cache[T any] interface {
	Set(key string, value T) error
	Get(key string) (T, error)
	Delete(key string) error
	Extract(key string) (T, error)
	Close() error
	Reset() error
}

type cache[T any] struct {
	status  uint32
	conf    *Config[T]
	buckets []bucket[T]
	null    T
}

func New[T any](conf *Config[T]) (Cache[T], error) {
	c := &cache[T]{
		status: cacheStatusActive,
		conf:   conf.Copy(),
	}

	if err := c.init(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *cache[T]) Set(key string, value T) error {
	if err := c.checkCache(cacheStatusActive); err != nil {
		return err
	}
	hkey := c.conf.Hasher.Sum64(key)
	b := &c.buckets[hkey%uint64(c.conf.Buckets)]
	return b.set(hkey, value)
}

func (c *cache[T]) Get(key string) (T, error) {
	if err := c.checkCache(cacheStatusActive); err != nil {
		return c.null, err
	}
	hkey := c.conf.Hasher.Sum64(key)
	b := &c.buckets[hkey%uint64(c.conf.Buckets)]
	return b.get(hkey)
}

func (c *cache[T]) Delete(key string) error {
	if err := c.checkCache(cacheStatusActive); err != nil {
		return err
	}
	hkey := c.conf.Hasher.Sum64(key)
	b := &c.buckets[hkey%uint64(c.conf.Buckets)]
	return b.delete(hkey)
}

func (c *cache[T]) Extract(key string) (T, error) {
	if err := c.checkCache(cacheStatusActive); err != nil {
		return c.null, err
	}
	hkey := c.conf.Hasher.Sum64(key)
	b := &c.buckets[hkey%uint64(c.conf.Buckets)]
	return b.extract(hkey)
}

func (c *cache[T]) Close() error {
	atomic.StoreUint32(&c.status, cacheStatusClosed)
	c.conf.Clock.Stop()
	return c.bulkClose()
}

func (c *cache[T]) Reset() error {
	return c.bulkReset()
}

func (c *cache[T]) bulkEvict() error {
	return c.bulkExec(c.conf.EvictWorkers, "eviction", func(b *bucket[T]) error {
		return b.evict()
	})
}

func (c *cache[T]) bulkClose() error {
	return c.bulkExec(c.conf.EvictWorkers, "close", func(b *bucket[T]) error {
		return b.close()
	})
}

func (c *cache[T]) bulkReset() error {
	return c.bulkExec(c.conf.EvictWorkers, "reset", func(b *bucket[T]) error {
		return b.reset()
	})
}

func (c *cache[T]) dump() error {
	if c.conf.DumpWriter == nil {
		return ErrOK
	}
	if err := c.bulkExec(c.conf.DumpWriteWorkers, "dump", func(b *bucket[T]) error { return b.dump() }); err != nil {
		return err
	}
	return c.conf.DumpWriter.Flush()
}

func (c *cache[T]) load() (int, error) {
	stream := make(chan Entry, c.conf.DumpReadBuffer)
	var wg sync.WaitGroup
	for i := uint(0); i < c.conf.DumpReadWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case e, ok := <-stream:
					if !ok {
						return
					}
					bkt := &c.buckets[e.Key%uint64(c.conf.Buckets)]
					var t T
					if err := c.conf.DumpDecoder.Decode(&t, e.Body); err != nil {
						continue
					}
					bkt.svcLock()
					_ = bkt.setLF(e.Key, t, int64(e.Expire))
					bkt.svcUnlock()
					c.mw().Load(bkt.id)
				}
			}
		}()
	}

	var lc int
	for {
		e, err := c.conf.DumpReader.Read()
		if err != nil {
			close(stream)
			if err != io.EOF && c.l() != nil {
				c.l().Printf("dump load interrupt due to error: %s", err.Error())
			}
			break
		}
		stream <- e
		lc++
	}

	wg.Wait()

	return lc, nil
}

func (c *cache[T]) bulkExec(workers uint, op string, fn func(b *bucket[T]) error) error {
	if workers == 0 || workers > c.conf.Buckets {
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
					bkt := &c.buckets[idx]
					if err := fn(bkt); err != nil && c.l() != nil {
						c.l().Printf("bucket #%d: %s failed with error '%s'\n", idx, op, err.Error())
					}
					continue
				}
				break
			}
		}(i)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := uint(0); i < c.conf.Buckets; i++ {
			bucketQueue <- i
		}
		close(bucketQueue)
	}()

	wg.Wait()

	return nil
}

func (c *cache[T]) checkCache(allow uint32) error {
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

func (c *cache[T]) mw() MetricsWriter {
	return c.conf.MetricsWriter
}

func (c *cache[T]) l() Logger {
	return c.conf.Logger
}

func (c *cache[T]) init() error {
	if c.conf == nil {
		return ErrNoConfig
	}
	if c.conf.Hasher == nil {
		return ErrNoHasher
	}
	if c.conf.Buckets == 0 {
		return ErrNoBuckets
	}

	if c.conf.MetricsWriter == nil {
		c.conf.MetricsWriter = dummyMW{}
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
			return ErrShortTTL
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

	if c.conf.DumpWriter != nil && c.conf.DumpEncoder != nil && c.conf.DumpInterval > 0 {
		if c.conf.DumpWriteWorkers == 0 {
			c.conf.DumpWriteWorkers = defaultDumpWriteWorkers
		}
		c.conf.Clock.Schedule(c.conf.DumpInterval, func() {
			if err := c.dump(); err != nil && c.l() != nil {
				c.l().Printf("dump write failed with error %s\n", err.Error())
			}
		})
	}

	if c.conf.DumpReader != nil && c.conf.DumpDecoder != nil {
		if c.conf.DumpReadWorkers == 0 {
			c.conf.DumpReadWorkers = defaultDumpReadWorkers
		}
		if c.conf.DumpReadBuffer == 0 {
			c.conf.DumpReadBuffer = c.conf.DumpReadWorkers
		}
		fn := func() {
			lc, err := c.load()
			if c.l() != nil {
				if err != nil {
					c.l().Printf("dump read failed with error %s\n", err.Error())
				} else {
					c.l().Printf("read %d entries from dump\n", lc)
				}
			}
		}
		if c.conf.DumpReadAsync {
			go fn()
		} else {
			fn()
		}
	}

	return nil
}
