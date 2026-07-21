package ttlcache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("copy", func(t *testing.T) {
		conf := Config[testEntry]{
			TTLInterval: time.Minute,
		}
		cpy := conf.Copy()
		conf.TTLInterval = 30 * time.Second
		assert.Equal(t, time.Minute, cpy.TTLInterval)
	})
}
