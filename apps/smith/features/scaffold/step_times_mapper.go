package scaffold

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// stepTime is how long one step of smith add took.
type stepTime struct {
	name string
	took time.Duration
}

// stepTimes collects the times of the steps of smith add, for the summary.
type stepTimes []stepTime

// Do runs a step and records its time under name, whether or not it fails.
func (t *stepTimes) Do(name string, step func() error) error {
	began := time.Now()
	err := step()
	*t = append(*t, stepTime{name, time.Since(began)})
	return err
}

// Line is the summary of the steps that took a visible time ("Steps: create
// 1.2s, install 14.3s"), or "" when none did.
func (t stepTimes) Line() string {
	var parts []string
	for _, s := range t {
		if s.took >= 100*time.Millisecond {
			parts = append(parts, fmt.Sprintf("%s %s", s.name, s.took.Round(100*time.Millisecond)))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "Steps: " + strings.Join(parts, ", ")
}

// heartbeat prints "still <what>... 10s" to out every interval until the
// returned function is called, so a long install shows that it is alive.
func heartbeat(out io.Writer, what string, interval time.Duration) (stop func()) {
	done := make(chan struct{})
	finished := make(chan struct{})
	began := time.Now()
	go func() {
		defer close(finished)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				fmt.Fprintf(out, "  still %s... %s\n", what, time.Since(began).Round(time.Second))
			}
		}
	}()
	return func() {
		close(done)
		<-finished
	}
}
