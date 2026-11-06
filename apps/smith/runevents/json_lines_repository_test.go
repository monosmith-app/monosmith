package runevents

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestJSONLinesWritesOneObjectPerLine(t *testing.T) {
	var out bytes.Buffer
	sink := NewJSONLines(&out)
	ok := true
	sink.Emit(Event{Event: TaskOutput, Run: "r1", Seq: 2, Time: time.Unix(0, 0).UTC(), Task: "web:build", Stream: Combined, Line: "hello"})
	sink.Emit(Event{Event: RunFinished, Run: "r1", Seq: 3, OK: &ok, Summary: &Summary{Passed: 1}})
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %q", out.String())
	}
	var first map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatal(err)
	}
	if first["event"] != "task.output" || first["task"] != "web:build" || first["line"] != "hello" || first["stream"] != "combined" {
		t.Fatalf("first event %v", first)
	}
	if _, has := first["status"]; has {
		t.Fatalf("fields of other kinds must be absent: %v", first)
	}
	if !strings.Contains(lines[1], `"ok":true`) {
		t.Fatalf("a finished run says ok even when true: %s", lines[1])
	}
}
