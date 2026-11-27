package rundurations

import (
	"testing"
	"time"
)

func TestTheEstimateSharesTheWorkLeftBetweenTheRunningTasks(t *testing.T) {
	h := History{"a": {10000}, "b": {20000}, "c": {6000}}
	left, ok := Estimate(h, []Pending{{ID: "a", Running: true, Elapsed: 4 * time.Second}, {ID: "b", Running: true}, {ID: "c"}})
	if !ok || left != 16*time.Second {
		t.Fatalf("(6s + 20s + 6s) / 2 running: %v %v", left, ok)
	}
	if left, _ := Estimate(h, []Pending{{ID: "a", Running: true, Elapsed: time.Minute}}); left != 0 {
		t.Fatalf("a task past its typical time leaves nothing: %v", left)
	}
}

func TestATaskWithoutHistoryGivesNoEstimate(t *testing.T) {
	if _, ok := Estimate(History{"a": {1000}}, []Pending{{ID: "a"}, {ID: "new"}}); ok {
		t.Fatal("one task without history is no estimate")
	}
	if _, ok := Estimate(History{}, nil); ok {
		t.Fatal("nothing pending is no estimate")
	}
}

func TestAnEstimateIsRounded(t *testing.T) {
	for in, want := range map[time.Duration]time.Duration{
		time.Second:                5 * time.Second,
		12400 * time.Millisecond:   12 * time.Second,
		150 * time.Second:          3 * time.Minute,
		61*time.Second + 100000000: time.Minute,
	} {
		if got := Round(in); got != want {
			t.Errorf("%v: %v, want %v", in, got, want)
		}
	}
}
