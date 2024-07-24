package mychannel

import (
	"context"
	"sync"
)

type FanOutHolder[T any] struct {
	mu      *sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	closed  bool

	in   <-chan T
	outs []chan T
}

func FanOut[T any](ctx context.Context, in chan T) FanOutHolder[T] {
	ctx, cancel := context.WithCancel(ctx)
	return FanOutHolder[T]{
		mu:     &sync.RWMutex{},
		ctx:    ctx,
		cancel: cancel,
		in:     in,
	}
}

// Start starts the fan-out process. It reads from the input channel and sends
// the value to all output channels. It will stop when the input channel is
// closed or the context is cancelled.
// Important: depending on the channel size, the order of the values may not be
// preserved. Especially when the channel size is 0, the value will be sent
// out of order if no one is reading from the channel.
func (h *FanOutHolder[T]) Start() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.started {
		return
	}
	if h.closed {
		panic("FanOutHolder already closed")
	}
	h.started = true

	go func() {
		defer h.Close()
		for {
			v, ok, ctxalive := ReadOne(h.ctx, h.in)
			if !ok || !ctxalive {
				return
			}
			h.mu.RLock()
			for _, ch := range h.outs {
				select {
				case ch <- v:
				default: // if no one is reading from the channel, send it later, out of order
					go func() {
						ch <- v
					}()
				}
			}
			h.mu.RUnlock()
		}
	}()
}

func (h *FanOutHolder[T]) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return
	}
	h.closed = true

	h.cancel()
	for _, ch := range h.outs {
		close(ch)
	}
}

func (h *FanOutHolder[T]) Add(chsize int) <-chan T {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		panic("FanOutHolder already closed")
	}
	if chsize < 0 {
		panic("channel size must be non-negative")
	}

	ch := make(chan T, chsize)

	h.outs = append(h.outs, ch)
	return ch
}

func (h *FanOutHolder[T]) Remove(ch <-chan T) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return
	}

	idx := -1
	for i, c := range h.outs {
		if c == ch {
			idx = i
			break
		}
	}
	if idx == -1 {
		return
	}
	Drain(h.outs[idx])
	h.outs = append(h.outs[:idx], h.outs[idx+1:]...)
}
