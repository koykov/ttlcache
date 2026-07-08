package ttlcache

const (
	offset = uint64(0xCBF29CE484222325)
	prime  = uint64(0x100000001B3)
)

// Simple FNV hasher implementation.
type testHasher struct{}

func (testHasher) Sum64(s string) uint64 {
	h := offset
	for i := 0; i < len(s); i++ {
		h *= prime
		h ^= uint64(s[i])
	}
	return h
}
