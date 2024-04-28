package ttlcache

import "time"

type MetricsWriter interface {
	Set(bucket string, dur time.Duration)
	Hit(bucket string, dur time.Duration)
	Del(bucket string)
	Miss(bucket string)
	Expire(bucket string)
	Overflow(bucket string)
	Evict(bucket string)
	Dump(bucket string)
	Load(bucket string)
}
