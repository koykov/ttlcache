package prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
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

func NewWriter(key string) *Writer {
	return NewWriterWP(key, time.Nanosecond)
}

func NewWriterWP(key string, precision time.Duration) *Writer {
	return &Writer{key: key, prec: precision}
}

func (w Writer) Set(bucket string, dur time.Duration) {
	size.WithLabelValues(w.key, bucket, cacheTotal).Inc()
	io.WithLabelValues(w.key, bucket, cacheIOSet).Inc()
	speed.WithLabelValues(w.key, bucket, speedWrite).Observe(float64(dur.Nanoseconds() / int64(w.prec)))
}

func (w Writer) Hit(bucket string, dur time.Duration) {
	io.WithLabelValues(w.key, bucket, cacheIOHit).Inc()
	speed.WithLabelValues(w.key, bucket, speedRead).Observe(float64(dur.Nanoseconds() / int64(w.prec)))
}

func (w Writer) Del(bucket string) {
	size.WithLabelValues(w.key, bucket, cacheDelete).Inc()
	io.WithLabelValues(w.key, bucket, cacheIODel).Inc()
}

func (w Writer) Miss(bucket string) {
	io.WithLabelValues(w.key, bucket, cacheIOMiss).Inc()
}

func (w Writer) Expire(bucket string) {
	io.WithLabelValues(w.key, bucket, cacheIOExpire).Inc()
}

func (w Writer) Overflow(bucket string) {
	io.WithLabelValues(w.key, bucket, cacheIONoSpace).Inc()
}

func (w Writer) Evict(bucket string) {
	size.WithLabelValues(w.key, bucket, cacheTotal).Dec()
	io.WithLabelValues(w.key, bucket, cacheIOEvict).Inc()
}

func (w Writer) Dump(bucket string) {
	dumpIO.WithLabelValues(w.key, bucket, dumpIODump).Inc()
}

func (w Writer) Load(bucket string) {
	dumpIO.WithLabelValues(w.key, bucket, dumpIOLoad).Inc()
}

var (
	size       *prometheus.GaugeVec
	io, dumpIO *prometheus.CounterVec
	speed      *prometheus.HistogramVec
)

func init() {
	size = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ttlcache_size",
		Help: "Total, used and free cache (bucket) size in bytes.",
	}, []string{"cache", "bucket", "type"})
	io = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ttlcache_io",
		Help: "Count cache IO operations calls.",
	}, []string{"cache", "bucket", "op"})

	dumpIO = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ttlcache_dump",
		Help: "Count dump IO operations calls.",
	}, []string{"cache", "bucket", "op"})

	speedBuckets := append(prometheus.DefBuckets, []float64{15, 20, 30, 40, 50, 100, 150, 200, 250, 500, 1000, 1500, 2000, 3000, 5000}...)
	speed = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ttlcache_io_speed",
		Help:    "Cache IO operations speed.",
		Buckets: speedBuckets,
	}, []string{"cache", "bucket", "op"})

	prometheus.MustRegister(size, io, dumpIO, speed)
}
