package mychannel

import (
	"context"
	"testing"
)

func TestReadOne(t *testing.T) {
	tc := []struct {
		name    string
		prepare func() (chan int, context.Context)

		expval      int
		expchok     bool
		expctxalive bool
	}{
		{
			name: "test with a value",
			prepare: func() (chan int, context.Context) {
				ch := make(chan int, 1)
				ch <- 1
				ctx := context.Background()
				return ch, ctx
			},
			expval:      1,
			expchok:     true,
			expctxalive: true,
		},
		{
			name: "test with a closed channel",
			prepare: func() (chan int, context.Context) {
				ch := make(chan int, 1)
				close(ch)
				ctx := context.Background()
				return ch, ctx
			},
			expval:      0,
			expchok:     false,
			expctxalive: true,
		},
		{
			name: "test with a cancelled context",
			prepare: func() (chan int, context.Context) {
				ch := make(chan int, 1)
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ch, ctx
			},
			expval:      0,
			expchok:     false,
			expctxalive: false,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ch, ctx := tt.prepare()
			val, chok, ctxAlive := ReadOne(ctx, ch)
			if val != tt.expval {
				t.Errorf("expected %d, got %d", tt.expval, val)
			}
			if chok != tt.expchok {
				t.Errorf("expected %t, got %t", tt.expchok, chok)
			}
			if ctxAlive != tt.expctxalive {
				t.Errorf("expected %t, got %t", tt.expctxalive, ctxAlive)
			}
		})
	}
}

func TestWriteOne(t *testing.T) {
	tc := []struct {
		name    string
		prepare func() (chan int, context.Context)

		expctxalive bool
	}{
		{
			name: "test with a value",
			prepare: func() (chan int, context.Context) {
				ch := make(chan int, 1)
				ctx := context.Background()
				return ch, ctx
			},
			expctxalive: true,
		},
		{
			name: "test with cancelled context",
			prepare: func() (chan int, context.Context) {
				ch := make(chan int, 1)
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ch, ctx
			},
			expctxalive: false,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ch, ctx := tt.prepare()
			ctxAlive := WriteOne(ctx, ch, 1)
			if ctxAlive != tt.expctxalive {
				t.Errorf("expected %t, got %t", tt.expctxalive, ctxAlive)
			}
		})
	}
}
