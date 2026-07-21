package dumpfs

import (
	"io"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	r, _ := NewReader("testdata/example.bin", WithOnEOF(KeepFile))
	for {
		e, err := r.Read()
		if err == io.EOF {
			break
		}
		exp := getTestBody(int(e.Key))
		assert.Equal(t, exp, e.Body)
		assert.NotEqual(t, math.MaxUint32, e.Expire)
	}
}
