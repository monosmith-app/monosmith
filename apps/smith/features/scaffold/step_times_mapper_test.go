package scaffold

import (
	"strings"
	"testing"
	"time"
)

func TestStepTimesLineNamesTheSlowSteps(t *testing.T) {
	steps := stepTimes{{"create", 1200 * time.Millisecond}, {"tools", 10 * time.Millisecond}, {"install", 14300 * time.Millisecond}}
	if got := steps.Line(); got != "Steps: create 1.2s, install 14.3s" {
		t.Fatalf("line = %q", got)
	}
	if got := (stepTimes{{"tools", time.Millisecond}}).Line(); got != "" {
		t.Fatalf("quick steps give no line: %q", got)
	}
}

func TestHeartbeatPrintsWhileTheStepRuns(t *testing.T) {
	var out strings.Builder
	stop := heartbeat(&out, "installing", 5*time.Millisecond)
	time.Sleep(30 * time.Millisecond)
	stop()
	if !strings.Contains(out.String(), "still installing...") {
		t.Fatalf("no heartbeat: %q", out.String())
	}
}
