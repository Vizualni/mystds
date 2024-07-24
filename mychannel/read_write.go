package mychannel

import "context"

// ReadOne reads one value from the channel while respecting the context.
func ReadOne[T any](ctx context.Context, in <-chan T) (val T, chOK bool, ctxAlive bool) {
	var t T
	select {
	case v, ok := <-in:
		return v, ok, true
	case <-ctx.Done():
		return t, false, false
	}
}

// WriteOne writes one value to the channel while respecting the context.
func WriteOne[T any](ctx context.Context, out chan<- T, val T) (ctxAlive bool) {
	select {
	case out <- val:
		return true
	case <-ctx.Done():
		return false
	}
}
