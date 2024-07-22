package myshutdown

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestSignalShutdown(t *testing.T) {
	type teststate struct {
		callbackCalled bool
		exitStatus     int
	}

	sigtrigger := func() {
		mypid := os.Getpid()
		syscall.Kill(mypid, syscall.SIGUSR1)
	}

	tt := []struct {
		name     string
		callback func()
		timeout  time.Duration
		assert   func(*testing.T, teststate)
	}{
		{
			name: "with graceful signal it should call the callback",
			assert: func(t *testing.T, state teststate) {
				if !state.callbackCalled {
					t.Error("Callback not called")
				}
				if state.exitStatus != 0 {
					t.Errorf("Exit status %d, expected 0", state.exitStatus)
				}
			},
		},
		{
			name:    "with slow signal it should exit with status 1",
			timeout: 2 * time.Second,
			callback: func() {
				time.Sleep(2 * time.Second)
			},
			assert: func(t *testing.T, state teststate) {
				if !state.callbackCalled {
					t.Error("Callback not called")
				}
				if state.exitStatus != 1 {
					t.Errorf("Exit status %d, expected 1", state.exitStatus)
				}
			},
		},
		{
			name:    "with another signal it should exit with status 2",
			timeout: 2*time.Second + 50*time.Millisecond,
			callback: func() {
				sigtrigger()
				time.Sleep(2 * time.Second)
			},
			assert: func(t *testing.T, state teststate) {
				if !state.callbackCalled {
					t.Error("Callback not called")
				}
				if state.exitStatus != 2 {
					t.Errorf("Exit status %d, expected 2", state.exitStatus)
				}
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			reset()
			defer reset()
			ctx := context.Background()
			state := teststate{}
			mycallback := func() {
				state.callbackCalled = true
				fmt.Println("Callback called")
				if tc.callback != nil {
					tc.callback()
				}
			}
			myexit := func(status int) {
				state.exitStatus = status
			}
			osExit = myexit
			RegisterSignals(ctx, syscall.SIGUSR1)
			GracefulTimeout(1 * time.Second)
			OnShutdown(mycallback)
			sigtrigger()

			if tc.timeout == 0 {
				tc.timeout = 50 * time.Millisecond
			}
			time.Sleep(tc.timeout)

			tc.assert(t, state)
		})
	}
}
