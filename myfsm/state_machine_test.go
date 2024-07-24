package myfsm

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestTransitioningSimpleState(t *testing.T) {
	ctx := context.Background()
	t.Run("simple state", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		var state Func
		cnt := 0
		state = func() Transitioner {
			if cnt == 1337 {
				return nil
			}
			cnt++
			return state
		}

		result, err := Start(ctx, state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cnt != 1337 {
			t.Fatalf("unexpected count: %d", cnt)
		}
		if result != nil {
			t.Fatalf("unexpected result: %v", result)
		}
	})

	t.Run("simple state with error", func(t *testing.T) {
		_, err := Start(ctx, Func(func() Transitioner {
			return Error(fmt.Errorf("test"))
		}))
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("simple state with context cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		cancel()
		callCount := 0
		increaseCallCount := Func(func() Transitioner {
			callCount++
			return nil
		})
		_, err := Start(ctx, Func(func() Transitioner {
			return increaseCallCount
		}))
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("unexpected error: %v", err)
		}
		if callCount != 0 {
			t.Fatalf("unexpected call count: %d", callCount)
		}
	})

	t.Run("return value", func(t *testing.T) {
		ret, err := Start(ctx, Func(func() Transitioner {
			return Return("baba")
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ret != "baba" {
			t.Fatalf("unexpected return value: %v", ret)
		}
	})
}
