package runview

import (
	"reflect"
	"testing"
	"time"
)

func TestParseLineReadsWhatSmithPrints(t *testing.T) {
	cases := []struct {
		line string
		want Event
	}{
		{"orders:build ▶ go build -trimpath .", Event{Kind: Started, Task: "orders:build", Text: "go build -trimpath ."}},
		{"kit:build ✓ 106ms", Event{Kind: Finish, Task: "kit:build", Status: Passed, Duration: 106 * time.Millisecond, Text: "106ms"}},
		{"orders:build ✓ 3.31s", Event{Kind: Finish, Task: "orders:build", Status: Passed, Duration: 3310 * time.Millisecond, Text: "3.31s"}},
		{"kit:test ✓ from cache", Event{Kind: Finish, Task: "kit:test", Status: Cached, Text: "from cache"}},
		{"web:test × exited with code 1", Event{Kind: Finish, Task: "web:test", Status: Failed, Text: "exited with code 1"}},
		{"sdk:build │ * Creating isolated environment", Event{Kind: Output, Task: "sdk:build", Text: "* Creating isolated environment"}},
		{"sdk:build │   - setuptools>=75", Event{Kind: Output, Task: "sdk:build", Text: "  - setuptools>=75"}},
		{"sdk:build │ ", Event{Kind: Output, Task: "sdk:build"}},
		{"web:serve ■ stopped", Event{Kind: Stopped, Task: "web:serve", Text: "stopped"}},
		{"web:build ! not cached: disk full", Event{Kind: Warned, Task: "web:build", Text: "not cached: disk full"}},
		{"  » web:test because kit:test did not succeed", Event{Kind: Skip, Task: "web:test", Status: Skipped, Text: "because kit:test did not succeed"}},
		{"  × kit:test exited with code 2", Event{Kind: Failure, Task: "kit:test", Status: Failed, Text: "exited with code 2"}},
		{"✓ 4/4 tasks succeeded (0 from cache) in 1.211s", Event{Kind: Summary, Done: 4, Total: 4, OK: true, Text: "✓ 4/4 tasks succeeded (0 from cache) in 1.211s"}},
		{"× 1/4 tasks succeeded (0 from cache) in 3s", Event{Kind: Summary, Done: 1, Total: 4, Text: "× 1/4 tasks succeeded (0 from cache) in 3s"}},
		{"affected: 2 files changed → 3 project(s)", Event{Kind: Plain, Text: "affected: 2 files changed → 3 project(s)"}},
		{"nothing to run", Event{Kind: Plain, Text: "nothing to run"}},
	}
	for _, c := range cases {
		if got := ParseLine(c.line); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%q\n got  %+v\n want %+v", c.line, got, c.want)
		}
	}
}

func TestParseLineMatchesColouredLinesOnTheirPlainText(t *testing.T) {
	if got := ParseLine("web:test │ \x1b[32mok\x1b[0m"); got.Kind != Output || got.Task != "web:test" || got.Text != "\x1b[32mok\x1b[0m" {
		t.Errorf("output keeps its colour: %+v", got)
	}
	if got := ParseLine("\x1b[32m✓\x1b[0m 4/4 tasks succeeded (0 from cache) in 1s"); got.Kind != Summary || got.Done != 4 || got.Text != "✓ 4/4 tasks succeeded (0 from cache) in 1s" {
		t.Errorf("a coloured summary is still the summary: %+v", got)
	}
	if got := ParseLine("\x1b[31m  × web:build exited with code 1\x1b[0m"); got.Kind != Failure || got.Task != "web:build" {
		t.Errorf("a coloured failure line: %+v", got)
	}
	if got := ParseLine("\x1b[33mnpm warn\x1b[0m deprecated"); got.Kind != Plain || got.Text != "\x1b[33mnpm warn\x1b[0m deprecated" {
		t.Errorf("a coloured plain line keeps its colour: %+v", got)
	}
}

func TestParseLineIgnoresACarriageReturn(t *testing.T) {
	if got := ParseLine("kit:build ✓ 1s\r"); got.Kind != Finish || got.Duration != time.Second {
		t.Fatalf("%+v", got)
	}
}

func TestATaskThatPrintsALookAlikeIsStillOutput(t *testing.T) {
	// The prefix is "<task> │ ", so anything after it is the task's own text.
	got := ParseLine("web:test │ × 3/4 tasks succeeded")
	if got.Kind != Output || got.Task != "web:test" {
		t.Fatalf("%+v", got)
	}
}
