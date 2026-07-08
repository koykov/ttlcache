package victoria

import (
	"time"

	"github.com/koykov/vmchain"
)

const (
	cacheTotal  = "total"
	cacheDelete = "delete"

	cacheIOSet     = "set"
	cacheIOEvict   = "evict"
	cacheIOMiss    = "miss"
	cacheIOHit     = "hit"
	cacheIODel     = "del"
	cacheIOExpire  = "expire"
	cacheIONoSpace = "no space"

	speedWrite = "write"
	speedRead  = "read"

	dumpIODump = "dump"
	dumpIOLoad = "load"
)

type Writer struct {
	key  string
	prec time.Duration
}

func NewWriter(key string, options ...Option) *Writer {
	w := &Writer{key: key}
	for _, fn := range options {
		fn(w)
	}
	if w.prec == 0 {
		w.prec = time.Nanosecond
	}
	return w
}

func (w Writer) Set(bucket string, dur time.Duration) {
	vmchain.Gauge("ttlcache_size", nil).
		WithLabel("cache", w.key).
		WithLabel("bucket", bucket).
		WithLabel("type", cacheTotal).Inc()
	vmchain.Counter("ttlcache_io").
		WithLabel("cache", w.key).
		WithLabel("bucket", bucket).
		WithLabel("op", cacheIOSet).Inc()
	vmchain.Histogram("ttlcache_io_speed").
		WithLabel("cache", w.key).
		WithLabel("bucket", bucket).
		WithLabel("op", speedWrite).Update(float64(dur.Nanoseconds() / int64(w.prec)))
}

func (w Writer) Hit(bucket string, dur time.Duration) {
	vmchain.Counter("ttlcache_io").
		WithLabel("cache", w.key).
		WithLabel("bucket", bucket).
		WithLabel("op", cacheIOHit).Inc()
	vmchain.Histogram("ttlcache_io_speed").
		WithLabel("cache", w.key).
		WithLabel("bucket", bucket).
		WithLabel("op", speedRead).Update(float64(dur.Nanoseconds() / int64(w.prec)))
}

func (w Writer) Del(bucket string) {
	vmchain.Gauge("ttlcache_size", nil).
		WithLabel("cache", w.key).
		WithLabel("bucket", bucket).
		WithLabel("type", cacheDelete).Inc()
	vmchain.Counter("ttlcache_io").
		WithLabel("cache", w.key).
		WithLabel("bucket", bucket).
		WithLabel("op", cacheIODel).Inc()
}

func (w Writer) Miss(bucket string) {
	vmchain.Counter("ttlcache_io").
		WithLabel("cache", w.key).
		WithLabel("bucket", bucket).
		WithLabel("op", cacheIOMiss).Inc()
}

func (w Writer) Expire(bucket string) {
	vmchain.Counter("ttlcache_io").
		WithLabel("cache", w.key).
		WithLabel("bucket", bucket).
		WithLabel("op", cacheIOExpire).Inc()
}

func (w Writer) Overflow(bucket string) {
	vmchain.Counter("ttlcache_io").
		WithLabel("cache", w.key).
		WithLabel("bucket", bucket).
		WithLabel("op", cacheIONoSpace).Inc()
}

func (w Writer) Evict(bucket string) {
	vmchain.Gauge("ttlcache_size", nil).
		WithLabel("cache", w.key).
		WithLabel("bucket", bucket).
		WithLabel("type", cacheTotal).Dec()
	vmchain.Counter("ttlcache_io").
		WithLabel("cache", w.key).
		WithLabel("bucket", bucket).
		WithLabel("op", cacheIOEvict).Inc()
}

func (w Writer) Dump(bucket string) {
	vmchain.Counter("ttlcache_dump_io").
		WithLabel("cache", w.key).
		WithLabel("bucket", bucket).
		WithLabel("op", dumpIODump).Inc()
}

func (w Writer) Load(bucket string) {
	vmchain.Counter("ttlcache_dump_io").
		WithLabel("cache", w.key).
		WithLabel("bucket", bucket).
		WithLabel("op", dumpIOLoad).Inc()
}

var _ = NewWriter
