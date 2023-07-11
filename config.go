package ttlcache

import "time"

type Config struct {
	Size          uint64
	Buckets       uint
	Hasher        Hasher
	TTLInterval   time.Duration
	EvictInterval time.Duration
	EvictWorkers  uint

	DumpWriter       DumpWriter
	DumpInterval     time.Duration
	DumpWriteWorkers uint

	DumpReader      DumpReader
	DumpReadBuffer  uint
	DumpReadWorkers uint
	DumpReadAsync   bool

	MetricsWriter MetricsWriter
	Clock         Clock
	Logger        Logger
}

func (c *Config) Copy() *Config {
	cpy := *c
	return &cpy
}
