package mychannel

import (
	"context"
	"testing"
)

func TestFanOut(t *testing.T) {
	ch := make(chan int)
	ctx := context.Background()
	go func() {
		ch <- 1
		ch <- 2
		ch <- 3
	}()
	h := FanOut(ctx, ch)
	defer h.Close()
	out1 := h.Add(3)
	out2 := h.Add(3)
	h.Start()

	val1 := <-out1
	val2 := <-out2
	if val1 != 1 {
		t.Errorf("expected 1, got %d", val1)
	}
	if val2 != 1 {
		t.Errorf("expected 1, got %d", val2)
	}

	val1 = <-out1
	val2 = <-out2
	if val1 != 2 {
		t.Errorf("expected 2, got %d", val1)
	}
	if val2 != 2 {
		t.Errorf("expected 2, got %d", val2)
	}
	h.Remove(out1)

	_, ok := <-out1
	if ok {
		t.Errorf("expected channel to be closed, but it is not")
	}
	val2 = <-out2
	if val2 != 3 {
		t.Errorf("expected 3, got %d", val2)
	}
	out3 := h.Add(1)
	ch <- 4

	val2 = <-out2
	val3 := <-out3
	if val2 != 4 {
		t.Errorf("expected 4, got %d", val2)
	}
	if val3 != 4 {
		t.Errorf("expected 4, got %d", val3)
	}
}
