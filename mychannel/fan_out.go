package mychannel

import (
	"sync"
	"time"
)

type FanOutHolder[T any] struct {
	mu      *sync.RWMutex
	started bool
	closed  bool

	in   <-chan T
	outs map[chan<- T]closechansignal

	WriteTimeout time.Duration
}

func FanOut[T any](in chan T, outs ...chan T) FanOutHolder[T] {
	chansMap := make(map[chan<- T]closechansignal, len(outs))
	for _, ch := range outs {
		chansMap[ch] = make(closechansignal)
	}
	return FanOutHolder[T]{
		mu:   &sync.RWMutex{},
		outs: chansMap,
	}
}

func (h *FanOutHolder[T]) start() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.started {
		return
	}
	if h.closed {
		panic("FanOutHolder already closed")
	}
	h.started = true
	for ch, iamclosed := range h.outs {
		h.write(ch, iamclosed)
	}
}

func (h *FanOutHolder[T]) write(ch chan<- T, iamclosed closechansignal) {
	go func() {
		for v := range h.in {
			select {
			case ch <- v:
			case <-iamclosed:
				return
			case <-time.After(h.WriteTimeout):
				panic("Write timeout in FanOutHolder")
			}
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
	for _, iamclosed := range h.outs {
		iamclosed <- struct{}{}
		close(iamclosed)

	}
}

func (h *FanOutHolder[T]) Start() {
	h.start()
}

// add
func (h *FanOutHolder[T]) Add(ch chan<- T) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		panic("FanOutHolder already closed")
	}
	_, ok := h.outs[ch]
	if ok {
		return
	}
	iamclosed := make(closechansignal)
	h.outs[ch] = iamclosed
	if h.started {
		h.write(ch, iamclosed)
	}
}

func (h *FanOutHolder[T]) Remove(ch chan<- T) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return
	}
	iamclosed, ok := h.outs[ch]
	if !ok {
		return
	}
	delete(h.outs, ch)
	iamclosed <- struct{}{}
	close(iamclosed)
}
