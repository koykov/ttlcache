package victoria

import "time"

type Option func(w *writer)

func WithPrecision(precision time.Duration) Option {
	return func(w *writer) {
		w.prec = precision
	}
}
