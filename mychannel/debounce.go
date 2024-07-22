package mychannel

import (
	"context"
	"time"
)

func DebounceAll[T any](ctx context.Context, in <-chan T, delay time.Duration) <-chan []T {
	ret := make(chan []T)
	timer := time.NewTimer(delay)

	go func() {
		defer close(ret)

		defer timer.Stop()

		first, ok, ctxAlive := ReadOne(ctx, in)
		if !ok || !ctxAlive {
			return
		}
		values := []T{first}
		timer.Reset(delay)
	loop:
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				break loop
			case v := <-in:
				values = append(values, v)
			}
		}
		if !WriteOne(ctx, ret, values) {
			return
		}
	}()

	return ret
}

func DebounceFirst[T any](ctx context.Context, in <-chan T, delay time.Duration) <-chan T {
	all := DebounceAll(ctx, in, delay)

	ret := make(chan T)

	go func() {
		defer close(ret)

		for {
			values, ok, ctxAlive := ReadOne(ctx, all)
			if !ok || !ctxAlive {
				return
			}
			if !WriteOne(ctx, ret, values[0]) {
				return
			}
		}
	}()

	return ret
}

func DebounceLast[T any](ctx context.Context, in <-chan T, delay time.Duration) <-chan T {
	all := DebounceAll(ctx, in, delay)

	ret := make(chan T)

	go func() {
		defer close(ret)
		for {
			values, ok, ctxAlive := ReadOne(ctx, all)
			if !ok || !ctxAlive {
				return
			}
			if !WriteOne(ctx, ret, values[len(values)-1]) {
				return
			}
		}
	}()

	return ret
}
