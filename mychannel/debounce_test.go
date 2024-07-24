package mychannel

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestDebounceAll(t *testing.T) {
	defdur := 10 * time.Millisecond
	tt := []struct {
		name   string
		filler func(in chan int)
		assert func(out <-chan []int)
	}{
		{
			name: "test with one element",
			filler: func(in chan int) {
				in <- 1
			},
			assert: func(out <-chan []int) {
				els := <-out
				if len(els) != 1 {
					t.Errorf("expected one element, got %d", len(els))
				}
				if len(out) != 0 {
					t.Errorf("expected no more elements, got %d", len(out))
				}
				if els[0] != 1 {
					t.Errorf("expected 1, got %d", els[0])
				}
			},
		},
		{
			name: "test with a few elements before the default duration",
			filler: func(in chan int) {
				for i := 0; i < 100; i++ {
					go func(i int) {
						in <- i
					}(i)
				}
			},
			assert: func(out <-chan []int) {
				els := <-out
				if len(els) != 100 {
					t.Errorf("expected 100 elements, got %d", len(els))
				}
				if len(out) != 0 {
					t.Errorf("expected no more elements, got %d", len(out))
				}
			},
		},
		{
			name: "test with a few elements after the default duration",
			filler: func(in chan int) {
				dump := func(num int) {
					var wg sync.WaitGroup
					wg.Add(num)
					for i := 0; i < num; i++ {
						go func(i int) {
							defer wg.Done()
							in <- num
						}(i)
					}
					wg.Wait()
				}
				dump(3)
				time.Sleep(2 * defdur)
				dump(5)
				time.Sleep(2 * defdur)
				dump(2)
			},
			assert: func(out <-chan []int) {
				els := <-out
				if len(els) != 3 {
					t.Errorf("expected 3 elements, got %d", len(els))
				}
				els = <-out
				if len(els) != 5 {
					t.Errorf("expected 5 elements, got %d", len(els))
				}
				els = <-out
				if len(els) != 2 {
					t.Errorf("expected 2 elements, got %d", len(els))
				}
			},
		},
		{
			name: "test with a few elements after the default duration with canceled context",
			filler: func(in chan int) {
				dump := func(num int) {
					var wg sync.WaitGroup
					wg.Add(num)
					for i := 0; i < num; i++ {
						go func(i int) {
							defer wg.Done()
							in <- num
						}(i)
					}
					wg.Wait()
				}
				dump(3)
				time.Sleep(2 * time.Second)
				dump(5)
			},
			assert: func(out <-chan []int) {
				els := <-out
				if len(els) != 3 {
					t.Errorf("expected 3 elements, got %d", len(els))
				}
				els = <-out
				if len(els) != 0 {
					t.Errorf("expected 0 elements, got %d", len(els))
				}
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
			defer cancel()
			mych := make(chan int, 1)
			go tc.filler(mych)
			out := DebounceAll(ctx, mych, defdur)
			tc.assert(out)
		})
	}
}

func TestDebounceFirst(t *testing.T) {
	defdur := 10 * time.Millisecond
	tt := []struct {
		name   string
		filler func(in chan int)
		assert func(out <-chan int)
	}{
		{
			name: "test with one element",
			filler: func(in chan int) {
				in <- 1
			},
			assert: func(out <-chan int) {
				el := <-out
				if el != 1 {
					t.Errorf("expected 1, got %d", el)
				}
			},
		},
		{
			name: "test with a few elements before the default duration",
			filler: func(in chan int) {
				for i := 3; i < 100; i++ {
					in <- i
				}
			},
			assert: func(out <-chan int) {
				el := <-out
				if el != 3 {
					t.Errorf("expected 3, got %d", el)
				}
			},
		},
		{
			name: "test with a few elements after the default duration",
			filler: func(in chan int) {
				in <- 1
				in <- 2
				in <- 3
				in <- 4
				time.Sleep(2 * defdur)
				in <- 10
				in <- 20
				in <- 30
				in <- 40
				time.Sleep(1 * time.Second)
				// this one is ignored
				in <- 100
			},
			assert: func(out <-chan int) {
				el := <-out
				if el != 1 {
					t.Errorf("expected 1, got %d", el)
				}
				el = <-out
				if el != 10 {
					t.Errorf("expected 10, got %d", el)
				}
				_, ok := <-out
				if ok {
					t.Errorf("expected channel to be closed, but it is not")
				}
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
			defer cancel()
			mych := make(chan int, 100)
			go tc.filler(mych)
			out := DebounceFirst(ctx, mych, defdur)
			tc.assert(out)
		})
	}
}

func TestDebounceLast(t *testing.T) {
	defdur := 10 * time.Millisecond
	tt := []struct {
		name   string
		filler func(in chan int)
		assert func(out <-chan int)
	}{
		{
			name: "test with one element",
			filler: func(in chan int) {
				in <- 1
			},
			assert: func(out <-chan int) {
				el := <-out
				if el != 1 {
					t.Errorf("expected 1, got %d", el)
				}
			},
		},
		{
			name: "test with a few elements before the default duration",
			filler: func(in chan int) {
				for i := 3; i <= 100; i++ {
					in <- i
				}
			},
			assert: func(out <-chan int) {
				el := <-out
				if el != 100 {
					t.Errorf("expected 100, got %d", el)
				}
			},
		},
		{
			name: "test with a few elements after the default duration",
			filler: func(in chan int) {
				in <- 1
				in <- 2
				in <- 3
				in <- 4
				time.Sleep(2 * defdur)
				in <- 10
				in <- 20
				in <- 30
				in <- 40
				time.Sleep(1 * time.Second)
				// this one is ignored
				in <- 100
			},
			assert: func(out <-chan int) {
				el := <-out
				if el != 4 {
					t.Errorf("expected 4, got %d", el)
				}
				el = <-out
				if el != 40 {
					t.Errorf("expected 40, got %d", el)
				}
				_, ok := <-out
				if ok {
					t.Errorf("expected channel to be closed, but it is not")
				}
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
			defer cancel()
			mych := make(chan int, 100)
			go tc.filler(mych)
			out := DebounceLast(ctx, mych, defdur)
			tc.assert(out)
		})
	}
}
