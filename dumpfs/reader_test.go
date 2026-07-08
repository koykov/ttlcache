package dumpfs

import (
	"bytes"
	"io"
	"math"
	"testing"

	"github.com/koykov/ttlcache"
)

func TestReader(t *testing.T) {
	r, _ := NewReader("testdata/example.bin", KeepFile)
	for {
		e, err := r.Read()
		if err == io.EOF {
			break
		}
		if !assertEntry(e) {
			t.FailNow()
		}
	}
}

func assertEntry(e ttlcache.Entry) bool {
	exp := getTestBody(int(e.Key))
	if !bytes.Equal(exp, e.Body) {
		return false
	}
	if e.Expire != math.MaxUint32 {
		return false
	}
	return true
}
