package victoria

import "time"

type Option func(w *Writer)

func WithPrecision(precision time.Duration) Option {
	return func(w *Writer) {
		w.prec = precision
	}
}
