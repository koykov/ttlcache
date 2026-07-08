package dumpfs

type WOption func(w *writer)

func WithBufferSize(bufferSize uint64) WOption {
	return func(w *writer) {
		w.bs = bufferSize
	}
}
