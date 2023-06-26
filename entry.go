package ttlcache

type entry[T any] struct {
	payload   T
	hkey      uint64
	timestamp int64
}
