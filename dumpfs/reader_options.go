package dumpfs

type ROption func(r *reader)

func WithOnEOF(onEOF OnEOF) ROption {
	return func(r *reader) {
		r.eof = onEOF
	}
}
