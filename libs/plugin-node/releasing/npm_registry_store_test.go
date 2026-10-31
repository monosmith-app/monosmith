package releasing

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestConfiguredRegistryAsksOncePerRootWithinTheTTL(t *testing.T) {
	now := time.Now()
	s := &registryStore{now: func() time.Time { return now }, entries: map[string]registryEntry{}}
	asked := 0
	ask := func() string { asked++; return "https://registry.example/" }

	for range 5 {
		if got := s.configured("/ws", ask); got != "https://registry.example/" {
			t.Fatalf("registry = %q", got)
		}
	}
	s.configured("/other", ask)
	if asked != 2 {
		t.Fatalf("asked %d times, want 2 (one per root)", asked)
	}

	now = now.Add(registryTTL + time.Second)
	s.configured("/ws", ask)
	if asked != 3 {
		t.Fatalf("asked %d times after the TTL, want 3", asked)
	}
}

func TestConfiguredRegistryIsSafeForConcurrentCallers(t *testing.T) {
	s := &registryStore{now: time.Now, entries: map[string]registryEntry{}}
	var wg sync.WaitGroup
	var asked atomic.Int32
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.configured("/ws", func() string { asked.Add(1); time.Sleep(10 * time.Millisecond); return "r" })
		}()
	}
	wg.Wait()
	if asked.Load() != 1 {
		t.Fatalf("asked %d times, want 1", asked.Load())
	}
}

func BenchmarkConfiguredRegistryRemembered(b *testing.B) {
	s := &registryStore{now: time.Now, entries: map[string]registryEntry{}}
	s.configured("/ws", func() string { return "r" })
	for b.Loop() {
		s.configured("/ws", func() string { return "r" })
	}
}
