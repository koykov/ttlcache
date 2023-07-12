package ttlcache

import "errors"

var (
	ErrNoConfig    = errors.New("no config provided")
	ErrBadCache    = errors.New("cache uninitialized, use New()")
	ErrCacheClosed = errors.New("cache closed")
	ErrNoHasher    = errors.New("no hasher provided")
	ErrNoBuckets   = errors.New("buckets must be greater than zero")
	ErrShortTTL    = errors.New("TTL must be greater than one second")
	ErrNotFound    = errors.New("entry not found")
	ErrExpire      = errors.New("entry expired")
)
