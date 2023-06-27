package ttlcache

import "time"

type Config struct {
	Size          uint64
	Buckets       uint
	Hasher        Hasher
	TTLInterval   time.Duration
	EvictInterval time.Duration
	EvictWorkers  uint
	Clock         Clock
	Logger        Logger
}

func (c *Config) Copy() *Config {
	cpy := *c
	return &cpy
}
