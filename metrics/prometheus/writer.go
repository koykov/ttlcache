package prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	cacheIOSet     = "set"
	cacheIOEvict   = "evict"
	cacheIOMiss    = "miss"
	cacheIOHit     = "hit"
	cacheIODelete  = "delete"
	cacheIOExtract = "extract"
	cacheIOExpire  = "expire"
	cacheIONoSpace = "no space"

	speedWrite = "write"
	speedRead  = "read"

	dumpIODump = "dump"
	dumpIOLoad = "load"
)

type Writer interface {
	Set(bucket string, dur time.Duration)
	Hit(bucket string, dur time.Duration)
	Delete(bucket string)
	Extract(bucket string)
	Miss(bucket string)
	Expire(bucket string)
	Overflow(bucket string)
	Evict(bucket string)
	Dump(bucket string)
	Load(bucket string)
}

type writer struct {
	key  string
	prec time.Duration
}

func NewWriter(key string, options ...Option) Writer {
	w := &writer{key: key}
	for _, fn := range options {
		fn(w)
	}
	if w.prec == 0 {
		w.prec = time.Nanosecond
	}
	return w
}

func (w *writer) Set(bucket string, dur time.Duration) {
	size.WithLabelValues(w.key, bucket).Inc()
	io.WithLabelValues(w.key, bucket, cacheIOSet).Inc()
	speed.WithLabelValues(w.key, bucket, speedWrite).Observe(float64(dur.Nanoseconds() / int64(w.prec)))
}

func (w *writer) Hit(bucket string, dur time.Duration) {
	io.WithLabelValues(w.key, bucket, cacheIOHit).Inc()
	speed.WithLabelValues(w.key, bucket, speedRead).Observe(float64(dur.Nanoseconds() / int64(w.prec)))
}

func (w *writer) Delete(bucket string) {
	size.WithLabelValues(w.key, bucket).Dec()
	io.WithLabelValues(w.key, bucket, cacheIODelete).Inc()
}

func (w *writer) Extract(bucket string) {
	size.WithLabelValues(w.key, bucket).Dec()
	io.WithLabelValues(w.key, bucket, cacheIOExtract).Inc()
}

func (w *writer) Miss(bucket string) {
	io.WithLabelValues(w.key, bucket, cacheIOMiss).Inc()
}

func (w *writer) Expire(bucket string) {
	io.WithLabelValues(w.key, bucket, cacheIOExpire).Inc()
}

func (w *writer) Overflow(bucket string) {
	io.WithLabelValues(w.key, bucket, cacheIONoSpace).Inc()
}

func (w *writer) Evict(bucket string) {
	size.WithLabelValues(w.key, bucket).Dec()
	io.WithLabelValues(w.key, bucket, cacheIOEvict).Inc()
}

func (w *writer) Dump(bucket string) {
	dumpIO.WithLabelValues(w.key, bucket, dumpIODump).Inc()
}

func (w *writer) Load(bucket string) {
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
	}, []string{"cache", "bucket"})
	io = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ttlcache_io",
		Help: "Count cache IO operations calls.",
	}, []string{"cache", "bucket", "op"})

	dumpIO = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ttlcache_dump_io",
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

var _ = NewWriter
