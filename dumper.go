package ttlcache

type Entry struct {
	Key    uint64
	Body   []byte
	Expire uint32
}

type DumpWriter interface {
	WriteVersion(version uint32) error
	Write(entry Entry) (int, error)
	Flush() error
}

type DumpReader interface {
	ReadVersion() uint32
	Read() (Entry, error)
}
