package dumpfs

import (
	"encoding/binary"
	"io"
	"os"
	"sync"

	"github.com/koykov/bytealg"
	"github.com/koykov/ttlcache"
)

type Reader interface {
	Read() (ttlcache.Entry, error)
}

type reader struct {
	fp  string
	eof func(string) error

	mux sync.Mutex
	f   *os.File
	buf []byte
}

func NewReader(filepath string, options ...ROption) (Reader, error) {
	r := &reader{fp: filepath}
	for _, fn := range options {
		fn(r)
	}
	if r.eof == nil {
		r.eof = os.Remove
	}
	return r, nil
}

func (r *reader) Read() (e ttlcache.Entry, err error) {
	r.mux.Lock()
	defer func() {
		err = r.checkEOF(err)
		r.mux.Unlock()
	}()

	if r.f == nil {
		if r.f, err = os.OpenFile(r.fp, os.O_RDONLY, 0644); err != nil {
			return
		}
	}

	r.buf = r.buf[:0]
	off := 0
	r.buf = bytealg.GrowDelta(r.buf, 8)
	if _, err = io.ReadAtLeast(r.f, r.buf[off:], 8); err != nil {
		return
	}
	e.Key = binary.LittleEndian.Uint64(r.buf[off:])
	off += 8

	var l int
	r.buf = bytealg.GrowDelta(r.buf, 4)
	if _, err = io.ReadAtLeast(r.f, r.buf[off:], 4); err != nil {
		return
	}
	l = int(binary.LittleEndian.Uint32(r.buf[off:]))
	off += 4
	blo := off
	r.buf = bytealg.GrowDelta(r.buf, l)
	if _, err = io.ReadAtLeast(r.f, r.buf[off:], l); err != nil {
		return
	}
	off += l
	bhi := off

	r.buf = bytealg.GrowDelta(r.buf, 4)
	if _, err = io.ReadAtLeast(r.f, r.buf[off:], 4); err != nil {
		return
	}
	e.Expire = binary.LittleEndian.Uint32(r.buf[off:])
	off += 4

	e.Body = append(e.Body, r.buf[blo:bhi]...)

	return
}

func (r *reader) checkEOF(err error) error {
	if err == io.EOF {
		_ = r.f.Close()
		_ = r.eof(r.fp)
		r.fp = ""
		r.f = nil
	}
	return err
}
