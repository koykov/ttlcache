package dumpfs

import "github.com/koykov/ttlcache"

type Writer struct{}

func (w Writer) WriteVersion(version uint32) error {
	// todo implement me
	return nil
}

func (w Writer) Write(entry ttlcache.Entry) (n int, err error) {
	_ = entry
	// todo implement me
	return
}

func (w Writer) Flush() (err error) {
	// todo implement me
	return
}
