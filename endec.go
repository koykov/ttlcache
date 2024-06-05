package ttlcache

type Encoder[T any] interface {
	Encode([]byte, T) ([]byte, int, error)
}

type Decoder[T any] interface {
	Decode(*T, []byte) error
}
