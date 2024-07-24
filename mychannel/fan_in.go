package mychannel

import (
	"context"
	"sync"
)

type closechansignal = chan struct{}

type FanInHolder[T any] struct {
	mu      *sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	closed  bool

	chans map[<-chan T]closechansignal
	out   chan T
}

// FanIn creates a new FanInHolder with the given channels. The standard
// pattern of using this as a function doesn't work well when you want to add
// or remove channels. This struct gives you more control over the whole E2E
// fan-in process.
func FanIn[T any](ctx context.Context, chans ...chan T) FanInHolder[T] {
	chansMap := make(map[<-chan T]closechansignal, len(chans))
	for _, ch := range chans {
		chansMap[ch] = make(closechansignal)
	}
	ctx, cancel := context.WithCancel(ctx)
	return FanInHolder[T]{
		ctx:    ctx,
		cancel: cancel,
		mu:     &sync.RWMutex{},
		chans:  chansMap,
		out:    make(chan T),
	}
}

func (h *FanInHolder[T]) start() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.started {
		return
	}
	if h.closed {
		panic("FanInHolder already closed")
	}
	h.started = true
	for ch, iamclosed := range h.chans {
		h.read(ch, iamclosed)
	}

	go func() {
		<-h.ctx.Done()
		h.Close()
	}()
}

func (h *FanInHolder[T]) read(ch <-chan T, iamclosed closechansignal) {
	go func() {
		for {
			select {
			case v, ok := <-ch:
				if !ok {
					return
				}
				h.out <- v
			case <-iamclosed:
				return
			}
		}
	}()
}

func (h *FanInHolder[T]) Chan() <-chan T {
	h.start()
	return h.out
}

func (h *FanInHolder[T]) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return
	}
	h.cancel()
	h.closed = true
	for _, iamclosed := range h.chans {
		iamclosed <- struct{}{}
		close(iamclosed)
	}
	close(h.out)
}

func (h *FanInHolder[T]) Remove(ch <-chan T) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return
	}
	iamclosed, ok := h.chans[ch]
	if !ok {
		return
	}
	delete(h.chans, ch)
	close(iamclosed)
}

func (h *FanInHolder[T]) Add(ch <-chan T) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return
	}
	if _, ok := h.chans[ch]; ok {
		return
	}
	h.chans[ch] = make(closechansignal)

	if h.started {
		h.read(ch, h.chans[ch])
	}
}
