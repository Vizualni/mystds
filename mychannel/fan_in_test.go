package mychannel

import (
	"context"
	"testing"
	"time"

	"github.com/vizualni/mystds/mytest"
)

func TestFanInReading(t *testing.T) {
	type testcase struct {
		chans    func() []chan int
		expected []int
	}
	tests := mytest.NewTests[testcase](t)

	tests.AddParallel("test with one channel", testcase{
		chans: func() []chan int {
			ch := make(chan int)
			go func() {
				ch <- 1
			}()
			return []chan int{ch}
		},
		expected: []int{1},
	})

	tests.AddParallel("test with multiple channels", testcase{
		chans: func() []chan int {
			ch1 := make(chan int)
			go func() {
				ch1 <- 1
			}()
			ch2 := make(chan int)
			go func() {
				ch2 <- 2
			}()
			ch3 := make(chan int)
			go func() {
				ch3 <- 3
			}()
			return []chan int{ch1, ch2, ch3}
		},
		expected: []int{1, 2, 3},
	})

	tests.Test(func(t *testing.T, tc testcase) {
		chans := tc.chans()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		holder := FanIn(ctx, chans...)

		received := make(map[int]struct{})
		for range len(chans) {
			val, _, ctxalive := ReadOne(holder.ctx, holder.Chan())
			if !ctxalive {
				t.Errorf("expected context to be alive, but it is not")
			}
			received[val] = struct{}{}
		}

		if len(received) != len(tc.expected) {
			t.Errorf("expected %d elements, got %d", len(tc.expected), len(received))
		}

		for _, val := range tc.expected {
			if _, ok := received[val]; !ok {
				t.Errorf("expected to receive %d, but it is not received", val)
			}
		}

		_, ok := <-holder.Chan()
		if ok {
			t.Errorf("expected channel to be closed, but it is not")
		}

		cancel()
	})
}

func TestFanInAddingAndRemovingChannels(t *testing.T) {
	t.Parallel()
	h := FanIn[int](context.Background())
	defer h.Close()
	ch := make(chan int, 1)
	ch <- 1
	h.Add(ch)
	hch := h.Chan()

	val, ok := <-hch
	if !ok {
		t.Errorf("expected channel to be open, but it is not")
	}
	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}

	newch := make(chan int, 1)
	h.Add(newch)

	// adding the same channel does nothing
	h.Add(newch)

	newch <- 2
	val, ok = <-hch
	if !ok {
		t.Errorf("expected channel to be open, but it is not")
	}
	if val != 2 {
		t.Errorf("expected 2, got %d", val)
	}
	h.Remove(ch)

	// removing the removed channel does nothing
	h.Remove(ch)

	newch <- 3
	ch <- 4

	val, ok = <-hch
	if !ok {
		t.Errorf("expected channel to be open, but it is not")
	}
	if val != 3 {
		t.Errorf("expected 3, got %d", val)
	}

	select {
	case val = <-hch:
		if val == 4 {
			t.Error("didn't expect any value from the removed channel")
		}
	default:
	}
	h.Close()

	Drain(ch)
	Drain(newch)

	// calling close again does nothing
	h.Close()
}
