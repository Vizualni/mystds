package myshutdown

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/vizualni/mystds/mychannel"
)

var (
	mu         *sync.RWMutex
	cs         []func()
	mainCtx    context.Context
	cancelFunc context.CancelFunc

	ch     chan struct{}
	_sigch chan os.Signal

	timeout = 30 * time.Second

	osExit              = os.Exit
	log    func(...any) = func(a ...any) { fmt.Println(a...) }

	closing = false
)

func init() {
	mu = &sync.RWMutex{}
	ch = make(chan struct{}, 1)

	reset()
	go func() {
		for range ch {
			TriggerShutdown()
		}
	}()
}

func reset() {
	mu.Lock()
	defer mu.Unlock()
	if _sigch != nil {
		signal.Stop(_sigch)
	}
	if cancelFunc != nil {
		cancelFunc()
	}
	cancelFunc = nil
	mainCtx = nil
	closing = false
	_sigch = make(chan os.Signal, 1)
	osExit = os.Exit
	log = func(a ...interface{}) { fmt.Println(a...) }
	timeout = 30 * time.Second
	cs = []func(){
		cancelMainContext,
	}
}

func SetLogger(l func(...interface{})) {
	log = l
}

func TriggerShutdown() {
	// not thread safe as we don't want to wait on the lock. The pattern is
	// that we only call this function from a signal handler or manually long
	// after the program has started, so it's safe to assume that the program
	// is not in a critical state.
	if closing {
		return
	}
	cancelFunc()
	closing = true

	allDone := make(chan struct{})

	// now we have to call all the shutdown callbacks
	var wg sync.WaitGroup
	for _, f := range cs {
		wg.Add(1)
		go func(f func()) {
			defer wg.Done()
			f()
		}(f)
	}
	go func() {
		wg.Wait()
		mychannel.Drain(ch)
	}()

	timer := time.NewTimer(timeout)
	// if we receive another signal, we have to exit immediately
	select {
	case <-allDone:
		log("All shutdown callbacks done, exiting")
		osExit(0)
	case <-timer.C:
		log("Graceful shutdown timed out, exiting immediately")
		osExit(1)
	case <-ch:
		log("Received another signal, exiting immediately")
		osExit(2)
	}
}

func GracefulTimeout(t time.Duration) {
	timeout = t
}

func Context() context.Context {
	mu.RLock()
	defer mu.RUnlock()
	if mainCtx == nil {
		panic("You must call OnSignals before Context")
	}
	return mainCtx
}

func OnShutdown(f func()) {
	mu.Lock()
	defer mu.Unlock()
	cs = append(cs, f)
}

func RegisterSignals(ctx context.Context, cs ...os.Signal) {
	mu.Lock()
	defer mu.Unlock()
	if cancelFunc != nil {
		panic("OnSignals must be called only once")
	}
	mainCtx, cancelFunc = context.WithCancel(ctx)
	_sigch = make(chan os.Signal, 1)
	signal.Notify(_sigch, cs...)
	go func() {
		defer mychannel.Drain(_sigch)
		for range _sigch {
			log("Received signal, triggering shutdown")
			ch <- struct{}{}
		}
	}()
}

func cancelMainContext() {
	mu.RLock()
	defer mu.RUnlock()
	if cancelFunc != nil {
		cancelFunc()
	}
}
