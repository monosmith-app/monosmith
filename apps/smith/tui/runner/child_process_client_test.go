package runner

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestMain lets the test binary act as the child: with RUNNER_HELPER set it
// prints some lines (and stderr too) and exits with the code in its last argument.
func TestMain(m *testing.M) {
	switch os.Getenv("RUNNER_HELPER") {
	case "":
		os.Exit(m.Run())
	case "lines":
		fmt.Println("kit:build ▶ go build")
		fmt.Fprintln(os.Stderr, "kit:build │ from stderr")
		fmt.Println("kit:build ✓ 1ms")
		if len(os.Args) > 0 && os.Args[len(os.Args)-1] == "fail" {
			os.Exit(3)
		}
		os.Exit(0)
	case "sleep":
		fmt.Println("started")
		time.Sleep(time.Minute)
	}
}

func helper(t *testing.T, mode string, extra ...string) *Process {
	t.Helper()
	self, _ := os.Executable()
	t.Setenv("RUNNER_HELPER", mode)
	p, err := Start(context.Background(), Spec{Program: self, Argv: append([]string{"-test.run=^$"}, extra...)})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func drain(p *Process) ([]string, Exit) {
	var all []string
	for {
		lines, exit := p.Next()
		all = append(all, lines...)
		if exit != nil {
			return all, *exit
		}
	}
}

func TestOutputFromBothStreamsArrivesAsLinesAndTheExitCodeIsKept(t *testing.T) {
	lines, exit := drain(helper(t, "lines", "fail"))
	if exit.Code != 3 || exit.Cancelled {
		t.Fatalf("exit %+v", exit)
	}
	text := strings.Join(lines, "\n")
	for _, want := range []string{"kit:build ▶ go build", "kit:build │ from stderr", "kit:build ✓ 1ms"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}
}

func TestASuccessfulChildExitsWithZero(t *testing.T) {
	_, exit := drain(helper(t, "lines"))
	if exit.Code != 0 || exit.Err != nil {
		t.Fatalf("exit %+v", exit)
	}
}

func TestStopEndsALongRunningChildAndSaysItWasCancelled(t *testing.T) {
	p := helper(t, "sleep")
	lines, _ := p.Next()
	if len(lines) == 0 || lines[0] != "started" {
		t.Fatalf("lines %v", lines)
	}
	p.Stop()
	done := make(chan Exit, 1)
	go func() { _, e := drain(p); done <- e }()
	select {
	case exit := <-done:
		if !exit.Cancelled {
			t.Fatalf("exit %+v", exit)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("the child was not stopped")
	}
}

func TestCommandLineIsWhatTheUserWouldType(t *testing.T) {
	if got := (Spec{Argv: []string{"run", "build", "-p", "kit"}}).CommandLine(); got != "smith run build -p kit" {
		t.Fatalf("%q", got)
	}
}

func TestTheChildGetsToolColourOnlyWhenAskedAndNoColourIsUnset(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	cases := []struct {
		name   string
		getenv func(string) string
		colour bool
		want   string
	}{
		{"a run pane", env(nil), true, "SMITH_INTERFACE=tui FORCE_COLOR=0 SMITH_TOOL_COLOR=1"},
		{"a capture the TUI reads", env(nil), false, "SMITH_INTERFACE=tui NO_COLOR=1 SMITH_TOOL_COLOR=0"},
		{"NO_COLOR wins", env(map[string]string{"NO_COLOR": "1"}), true, "SMITH_INTERFACE=tui NO_COLOR=1 SMITH_TOOL_COLOR=0"},
	}
	for _, c := range cases {
		if got := strings.Join(childEnv(c.getenv, c.colour), " "); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestStartingAMissingProgramFails(t *testing.T) {
	if _, err := Start(context.Background(), Spec{Program: "definitely-not-a-program-xyz"}); err == nil {
		t.Fatal("expected an error")
	}
}
