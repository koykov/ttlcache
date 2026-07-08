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

type dummyMW struct{}

func (dummyMW) Set(_ string, _ time.Duration) {}
func (dummyMW) Hit(_ string, _ time.Duration) {}
func (dummyMW) Del(_ string)                  {}
func (dummyMW) Miss(_ string)                 {}
func (dummyMW) Expire(_ string)               {}
func (dummyMW) Overflow(_ string)             {}
func (dummyMW) Evict(_ string)                {}
func (dummyMW) Dump(_ string)                 {}
func (dummyMW) Load(_ string)                 {}
