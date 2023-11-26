package ttlcache

type Encoder[T any] interface {
	Encode([]byte, T) ([]byte, int, error)
}

type Decode[T any] interface {
	Decode(*T, []byte) error
}
