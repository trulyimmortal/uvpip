package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"
)

// The readiness output proves that cancellation happens after process startup.
func runCancellationFixture() {
	fmt.Fprint(os.Stdout, "ready")
	time.Sleep(10 * time.Second) // Bound the fixture even if cancellation regresses.
}

type cancelOnReady struct {
	cancel context.CancelFunc
	ready  bool
}

func (w *cancelOnReady) Write(p []byte) (int, error) {
	if len(p) > 0 {
		w.ready = true
		w.cancel()
	}
	return len(p), nil
}

func TestRunUVRunningChildCancellation(t *testing.T) {
	for _, tt := range []struct {
		name        string
		beforeStart bool
	}{
		{"before startup", true},
		{"after readiness", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fixtureUV(t)
			t.Setenv("UVPIP_TEST_MODE", "wait-for-cancel")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			out := &cancelOnReady{cancel: cancel}
			if tt.beforeStart {
				cancel()
			}
			started := time.Now()
			code := runUVContext(ctx, []string{"pip", "list"}, nil, out, io.Discard, quiet)
			if code == exitOK || out.ready == tt.beforeStart {
				t.Fatalf("exit=%d ready=%v beforeStart=%v", code, out.ready, tt.beforeStart)
			}
			if !tt.beforeStart && code == exitCannotExecute {
				t.Fatal("running child was incorrectly reported as a startup failure")
			}
			if elapsed := time.Since(started); elapsed >= 5*time.Second {
				t.Fatalf("cancellation did not promptly reap child: %s", elapsed)
			}
		})
	}
}
