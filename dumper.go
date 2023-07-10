package ttlcache

type Entry struct {
	Key    uint64
	Body   []byte
	Expire uint32
}

type DumpWriter interface {
	Write(entry Entry) (int, error)
	Flush() error
}

type DumpReader interface {
	Read() (Entry, error)
}
