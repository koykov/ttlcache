package ttlcache

import "time"

type Config struct {
	Size           uint64
	Buckets        uint
	Hasher         Hasher
	ExpireInterval time.Duration
}

func (c *Config) Copy() *Config {
	cpy := *c
	return &cpy
}
