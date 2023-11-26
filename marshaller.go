package ttlcache

type MarshallerTo interface {
	Size() int
	MarshalTo([]byte) (int, error)
}

type Unmarshaller interface {
	Unmarshal([]byte) error
}
