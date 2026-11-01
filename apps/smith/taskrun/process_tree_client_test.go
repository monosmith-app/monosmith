//go:build !windows

package taskrun

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

// TestCancelStopsTheGrandchild checks that cancelling stops the whole tree,
// also when the child ignores SIGTERM and needs the forced kill.
func TestCancelStopsTheGrandchild(t *testing.T) {
	old := KillGrace
	KillGrace = 300 * time.Millisecond
	defer func() { KillGrace = old }()
	for name, script := range map[string]string{
		"sigterm":         "sleep 100 & echo $!; wait",
		"ignores sigterm": "trap '' TERM; sleep 100 & echo $!; wait",
	} {
		t.Run(name, func(t *testing.T) {
			out := &lockedBuffer{}
			ctx, cancel := context.WithCancel(context.Background())
			finished := make(chan struct{})
			go func() {
				defer close(finished)
				_, _ = RunProcess(ctx, Spec{Argv: []string{"sh", "-c", script}, Dir: t.TempDir(), Out: out})
			}()
			var pid int
			for deadline := time.Now().Add(5 * time.Second); pid == 0 && time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
				pid, _ = strconv.Atoi(strings.TrimSpace(out.String()))
			}
			if pid == 0 {
				t.Fatal("the grandchild did not start")
			}
			cancel()
			select {
			case <-finished:
			case <-time.After(10 * time.Second):
				t.Fatal("RunProcess did not return after cancel")
			}
			for deadline := time.Now().Add(3 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
				if syscall.Kill(pid, 0) != nil {
					return
				}
			}
			_ = syscall.Kill(pid, syscall.SIGKILL)
			t.Fatalf("grandchild %d is still running after cancel", pid)
		})
	}
}
