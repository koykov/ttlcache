package ttlcache

import "time"

type DummyMetrics struct{}

func (DummyMetrics) Set(_ string, _ time.Duration) {}
func (DummyMetrics) Hit(_ string, _ time.Duration) {}
func (DummyMetrics) Del(_ string)                  {}
func (DummyMetrics) Miss(_ string)                 {}
func (DummyMetrics) Expire(_ string)               {}
func (DummyMetrics) Overflow(_ string)             {}
func (DummyMetrics) Evict(_ string)                {}
func (DummyMetrics) Dump(_ string)                 {}
func (DummyMetrics) Load(_ string)                 {}
