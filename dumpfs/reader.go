package dumpfs

import "github.com/koykov/ttlcache"

type Reader struct{}

func (r Reader) ReadVersion() uint32 {
	// todo implement me
	return 0
}

func (r Reader) Read() (e ttlcache.Entry, err error) {
	// todo implement me
	return
}
