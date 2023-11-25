package dumpfs

import "github.com/koykov/ttlcache"

type Reader struct{}

func (r Reader) Read() (e ttlcache.Entry, err error) {
	// todo implement me
	return
}
