package dumpfs

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/koykov/bytealg"
	"github.com/koykov/byteconv"
	"github.com/koykov/clock"
	"github.com/koykov/ttlcache"
)

const defaultBlockSIze = 4096

type Writer struct {
	Buffer   uint64
	FilePath string

	once sync.Once
	bs   uint64
	fp   string
	fd   string
	ft   string
	bsz  int64

	mux sync.Mutex
	f   *os.File
	buf []byte

	err error
}

func (w *Writer) Write(entry ttlcache.Entry) (n int, err error) {
	w.once.Do(w.init)
	if w.err != nil {
		err = w.err
		return
	}

	w.mux.Lock()
	defer w.mux.Unlock()

	off := len(w.buf)
	poff := off
	w.buf = bytealg.GrowDelta(w.buf, 8)
	binary.LittleEndian.PutUint64(w.buf[off:], entry.Key)
	off += 8

	w.buf = bytealg.GrowDelta(w.buf, 4)
	binary.LittleEndian.PutUint32(w.buf[off:], uint32(len(entry.Body)))
	off += 4
	w.buf = append(w.buf, entry.Body...)
	off += len(entry.Body)

	w.buf = bytealg.GrowDelta(w.buf, 4)
	binary.LittleEndian.PutUint32(w.buf[off:], entry.Expire)
	off += 4

	n = off - poff

	if uint64(len(w.buf)) >= w.bs {
		err = w.flushBuf()
	}

	return
}

func (w *Writer) Flush() (err error) {
	w.once.Do(w.init)
	if w.err != nil {
		return w.err
	}

	w.mux.Lock()
	defer w.mux.Unlock()

	if len(w.buf) > 0 {
		if err = w.flushBuf(); err != nil {
			return
		}
	}

	if err = w.f.Close(); err != nil {
		return
	}
	err = os.Rename(w.ft, w.fd)
	w.f = nil

	return
}

func (w *Writer) init() {
	w.err = nil
	if len(w.FilePath) == 0 {
		w.err = ErrNoFilePath
		return
	}
	dir := filepath.Dir(w.FilePath)
	if !isDirWR(dir) {
		w.err = ErrDirNoWR
		return
	}
	if w.bsz = blockSizeOf(dir); w.bsz == 0 {
		w.bsz = defaultBlockSIze
	}

	w.fp = w.FilePath
	if w.bs = w.Buffer; w.bs > 0 {
		w.buf = make([]byte, 0, w.bs)
	}
}

func (w *Writer) flushBuf() (err error) {
	if w.f == nil {
		buf := make([]byte, 0, len(w.fp)*2)
		if buf, err = clock.AppendFormat(buf, w.fp, time.Now()); err != nil {
			return
		}
		w.fd = byteconv.B2S(buf)
		w.ft = w.fd + ".tmp"
		if w.f, err = os.Create(w.ft); err != nil {
			return
		}
	}

	p := w.buf
	for len(p) >= int(w.bsz) {
		if _, err = w.f.Write(p[:w.bsz]); err != nil {
			return
		}
		p = p[w.bsz:]
	}
	if len(p) > 0 {
		if _, err = w.f.Write(p); err != nil {
			return
		}
	}
	w.buf = w.buf[:0]
	return
}
