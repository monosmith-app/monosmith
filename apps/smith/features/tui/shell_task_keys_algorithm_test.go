package tui

import "testing"

func TestEachTaskTakesTheFirstFreeLetterOfItsName(t *testing.T) {
	taken := func(k string) bool { return k == "p" || k == "s" } // the view uses p and s
	got := chooseTaskKeys([]string{"package", "slice-check", "start", "serve"}, taken)
	want := map[string]string{"package": "a", "slice-check": "l", "start": "t", "serve": "e"}
	for task, key := range want {
		if got[task] != key {
			t.Errorf("%s: key %q, want %q (all: %v)", task, got[task], key, got)
		}
	}
}

func TestNoTwoTasksShareAKey(t *testing.T) {
	got := chooseTaskKeys([]string{"aa", "ab", "ac", "ad"}, func(string) bool { return false })
	seen := map[string]string{}
	for task, key := range got {
		if other, dup := seen[key]; dup {
			t.Fatalf("%s and %s share %q", task, other, key)
		}
		seen[key] = task
	}
}

func TestATaskWithNoFreeKeyGetsNone(t *testing.T) {
	if got := chooseTaskKeys([]string{"x"}, func(string) bool { return true }); got["x"] != "" {
		t.Fatalf("got %q", got["x"])
	}
}
