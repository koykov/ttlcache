package ttlcache

import "time"

type MetricsWriter interface {
	Set(bucket string, dur time.Duration)
	Hit(bucket string, dur time.Duration)
	Miss(bucket string)
	Expire(bucket string)
	Overflow(bucket string)
	Evict(bucket string)
}
