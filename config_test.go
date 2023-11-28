package ttlcache

import (
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	t.Run("copy", func(t *testing.T) {
		conf := Config[testEntry]{
			TTLInterval: time.Minute,
		}
		cpy := conf.Copy()
		conf.TTLInterval = 30 * time.Second
		if cpy.TTLInterval != time.Minute {
			t.Error("config copy failed")
		}
	})
}
