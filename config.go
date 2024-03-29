package ttlcache

import "time"

type Config[T any] struct {
	Size          uint64
	Buckets       uint
	Hasher        Hasher
	TTLInterval   time.Duration
	EvictInterval time.Duration
	EvictWorkers  uint

	DumpVersion uint32

	DumpWriter       DumpWriter
	DumpEncoder      Encoder[T]
	DumpInterval     time.Duration
	DumpWriteWorkers uint

	DumpReader      DumpReader
	DumpDecoder     Decode[T]
	DumpReadBuffer  uint
	DumpReadWorkers uint
	DumpReadAsync   bool

	MetricsWriter MetricsWriter
	Clock         Clock
	Logger        Logger
}

func (c *Config[T]) Copy() *Config[T] {
	cpy := *c
	return &cpy
}
