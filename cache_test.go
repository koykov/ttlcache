package ttlcache

import (
	"testing"
	"time"

	"github.com/koykov/fastconv"
	"github.com/koykov/hash/fnv"
)

type testEntry struct {
	p []byte
}

func TestCache(t *testing.T) {
	t.Run("delete", func(t *testing.T) {
		cache, err := New[testEntry](&Config{
			Buckets:     4,
			Hasher:      &fnv.Hasher{},
			TTLInterval: time.Minute,
		})
		if err != nil {
			t.Error(err)
		}
		if err = cache.Set("foo", testEntry{p: []byte("foobar")}); err != nil {
			t.Error(err)
		}
		if err = cache.Delete("foo"); err != nil {
			t.Error(err)
		}
		if err = cache.Close(); err != nil {
			t.Error(err)
		}
	})
}

func TestIO(t *testing.T) {
	testIO := func(t *testing.T, entries int, verbose bool) {
		cache, err := New[testEntry](&Config{
			Buckets:     4,
			Hasher:      &fnv.Hasher{},
			TTLInterval: time.Minute,
		})
		if err != nil {
			t.Fatal(err)
		}

		var (
			key []byte
			dst any

			w, wf, r, rf, r404 int
		)

		for i := 0; i < entries; i++ {
			w++
			key = makeKey(key, i)
			if err := cache.Set(fastconv.B2S(key), testEntry{p: getEntryBody(i)}); err != nil {
				wf++
				t.Error(err)
			}
		}

		for i := 0; i < entries; i++ {
			r++
			key = makeKey(key, i)
			if dst, err = cache.Get(fastconv.B2S(key)); err != nil {
				rf++
				r404++
				if err != ErrNotFound {
					r404--
					t.Error(err)
				}
				continue
			}
			assertBytes(t, getEntryBody(i), dst.(testEntry).p)
		}

		if verbose {
			t.Logf("write: %d\nwrite fail: %d\nread: %d\nread fail: %d\nread 404: %d", w, wf, r, rf, r404)
		}

		if err = cache.Close(); err != nil {
			t.Error(err)
		}
	}

	const verbose = false
	t.Run("1", func(t *testing.T) { testIO(t, 1, verbose) })
	t.Run("10", func(t *testing.T) { testIO(t, 10, verbose) })
	t.Run("100", func(t *testing.T) { testIO(t, 100, verbose) })
	t.Run("1K", func(t *testing.T) { testIO(t, 1000, verbose) })
	t.Run("10K", func(t *testing.T) { testIO(t, 10000, verbose) })
	t.Run("100K", func(t *testing.T) { testIO(t, 100000, verbose) })
	t.Run("1M", func(t *testing.T) { testIO(t, 1000000, verbose) })
}
