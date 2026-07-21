package ttlcache

import (
	"testing"
	"time"

	"github.com/koykov/byteconv"
	"github.com/stretchr/testify/assert"
)

type testEntry struct {
	p []byte
}

func TestCache(t *testing.T) {
	t.Run("delete", func(t *testing.T) {
		cache, err := New[testEntry](&Config[testEntry]{
			Buckets:     4,
			Hasher:      testHasher{},
			TTLInterval: time.Minute,
		})
		assert.NoError(t, err)
		err = cache.Set("foo", testEntry{p: []byte("foobar")})
		assert.NoError(t, err)
		err = cache.Delete("foo")
		assert.NoError(t, err)
		err = cache.Close()
		assert.NoError(t, err)
	})
	t.Run("extract", func(t *testing.T) {
		cache, err := New[testEntry](&Config[testEntry]{
			Buckets:     4,
			Hasher:      testHasher{},
			TTLInterval: time.Minute,
		})
		assert.NoError(t, err)
		err = cache.Set("foo", testEntry{p: []byte("foobar")})
		assert.NoError(t, err)
		_, err = cache.Extract("foo")
		assert.NoError(t, err)
		_, err = cache.Get("foo")
		assert.ErrorIs(t, err, ErrNotFound)
		err = cache.Close()
		assert.NoError(t, err)
	})
	t.Run("reset", func(t *testing.T) {
		cache, err := New[testEntry](&Config[testEntry]{
			Buckets:     4,
			Hasher:      testHasher{},
			TTLInterval: time.Minute,
		})
		assert.NoError(t, err)
		err = cache.Set("foo", testEntry{p: []byte("foobar")})
		assert.NoError(t, err)
		err = cache.Reset()
		assert.NoError(t, err)
		_, err = cache.Get("foo")
		assert.ErrorIs(t, err, ErrNotFound)
		err = cache.Close()
		assert.NoError(t, err)
	})
	t.Run("close", func(t *testing.T) {
		cache, err := New[testEntry](&Config[testEntry]{
			Buckets:     4,
			Hasher:      testHasher{},
			TTLInterval: time.Minute,
		})
		assert.NoError(t, err)
		err = cache.Set("foo", testEntry{p: []byte("foobar")})
		assert.NoError(t, err)
		err = cache.Close()
		assert.NoError(t, err)
		_, err = cache.Get("foo")
		assert.ErrorIs(t, err, ErrCacheClosed)
	})
}

func TestIO(t *testing.T) {
	testIO := func(t *testing.T, entries int, verbose bool) {
		cache, err := New[testEntry](&Config[testEntry]{
			Buckets:     4,
			Hasher:      testHasher{},
			TTLInterval: time.Minute,
		})
		assert.NoError(t, err)

		var (
			key []byte
			dst any

			w, wf, r, rf, r404 int
		)

		for i := 0; i < entries; i++ {
			w++
			key = makeKey(key, i)
			if err := cache.Set(byteconv.B2S(key), testEntry{p: getEntryBody(i)}); err != nil {
				wf++
				t.Error(err)
			}
		}

		for i := 0; i < entries; i++ {
			r++
			key = makeKey(key, i)
			if dst, err = cache.Get(byteconv.B2S(key)); err != nil {
				rf++
				r404++
				if err != ErrNotFound {
					r404--
					t.Error(err)
				}
				continue
			}
			assert.Equal(t, getEntryBody(i), dst.(testEntry).p)
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
