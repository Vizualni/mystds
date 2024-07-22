package mychannel

import "sync"

type closechansignal = chan struct{}

type FanInHolder[T any] struct {
	mu      *sync.RWMutex
	started bool
	closed  bool

	chans map[<-chan T]closechansignal
	out   chan T
}

func FanIn[T any](chans ...chan T) FanInHolder[T] {
	chansMap := make(map[<-chan T]closechansignal, len(chans))
	for _, ch := range chans {
		chansMap[ch] = make(closechansignal)
	}
	return FanInHolder[T]{
		mu:    &sync.RWMutex{},
		chans: chansMap,
		out:   make(chan T),
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
	h.chans[ch] = make(closechansignal)

	if h.started {
		h.read(ch, h.chans[ch])
	}
}
