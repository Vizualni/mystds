package mychannel

import (
	"sync"
	"testing"
)

func TestDraining(t *testing.T) {
	ch := make(chan int, 1024)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ch <- i
		}(i)
	}
	wg.Wait()

	if len(ch) == 0 {
		t.Error("expected channel to have elements, but it has none")
	}

	Drain(ch)
	if len(ch) != 0 {
		t.Errorf("expected channel to be drained, but it has %d elements", len(ch))
	}
	if !isChannelClosed(ch) {
		t.Error("expected channel to be closed, but it is not")
	}
}

func isChannelClosed[T any](ch chan T) bool {
	select {
	case _, ok := <-ch:
		return !ok
	default:
		return false
	}
}
