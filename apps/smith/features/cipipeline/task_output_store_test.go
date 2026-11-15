package cipipeline

import (
	"strings"
	"testing"
)

// The verify log may carry tool colour; what smith reads from it (the tail it
// prints outside the group, the test counts in the table and the job summary)
// stays plain.
func TestTaskOutputsReadTheLogWithoutItsColour(t *testing.T) {
	outputs := newTaskOutputs()
	_, _ = outputs.Write([]byte("core:test │ \x1b[32m====== 42 passed in 3s ======\x1b[0m\ncore:test │ \x1b[31m--- FAIL: TestX\x1b[0m\x1b[2K\n"))
	if summary, passed := outputs.Tests("core:test"); passed != 42 || strings.Contains(summary, "\x1b") {
		t.Errorf("test counts: %q, %d", summary, passed)
	}
	if tail := outputs.Tail("core:test"); strings.Contains(tail, "\x1b") || !strings.Contains(tail, "--- FAIL: TestX") {
		t.Errorf("tail: %q", tail)
	}
}
