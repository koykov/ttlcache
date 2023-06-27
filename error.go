package ttlcache

import "errors"

var (
	ErrNoConfig  = errors.New("no config provided")
	ErrNoHasher  = errors.New("no hasher provided")
	ErrNoBuckets = errors.New("buckets must be greater than zero")
)
