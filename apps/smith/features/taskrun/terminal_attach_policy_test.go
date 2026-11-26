package taskrun

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/taskgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// attachGraph is a graph of tasks by ID; requested are the ones asked for.
func attachGraph(tasks map[string]protocol.Task, requested ...string) *taskgraph.Graph {
	g := &taskgraph.Graph{Tasks: map[string]*taskgraph.Task{}}
	asked := map[string]bool{}
	for _, id := range requested {
		asked[id] = true
	}
	for _, id := range []string{"lib:build", "app:start", "app:build", "web:start"} {
		def, ok := tasks[id]
		if !ok {
			continue
		}
		g.Tasks[id] = &taskgraph.Task{ID: id, Def: def, Requested: asked[id]}
		g.Order = append(g.Order, id)
	}
	return g
}

func TestAttachedTaskIsTheOneUncachedTaskAskedFor(t *testing.T) {
	g := attachGraph(map[string]protocol.Task{"lib:build": {Cache: true}, "app:start": {}}, "app:start")
	if got := AttachedTask(g, true, false); got != "app:start" {
		t.Fatalf("got %q, want app:start: its dependency runs first, prefixed", got)
	}
}

func TestAttachedTaskNeedsATerminal(t *testing.T) {
	g := attachGraph(map[string]protocol.Task{"app:start": {}}, "app:start")
	if got := AttachedTask(g, false, false); got != "" {
		t.Fatalf("no terminal, got %q", got)
	}
	if got := AttachedTask(g, true, true); got != "" {
		t.Fatalf("a quiet run holds the output, got %q", got)
	}
}

func TestAttachedTaskSkipsSeveralTasks(t *testing.T) {
	g := attachGraph(map[string]protocol.Task{"app:start": {}, "web:start": {}}, "app:start", "web:start")
	if got := AttachedTask(g, true, false); got != "" {
		t.Fatalf("two tasks share one terminal, got %q", got)
	}
}

func TestAttachedTaskSkipsWhatNeedsTheOutput(t *testing.T) {
	cases := map[string]protocol.Task{
		"a cached task": {Cache: true},
		"a served app":  {Serve: "app:start"},
	}
	for name, def := range cases {
		g := attachGraph(map[string]protocol.Task{"app:build": def}, "app:build")
		if got := AttachedTask(g, true, false); got != "" {
			t.Errorf("%s: got %q", name, got)
		}
	}
	g := attachGraph(map[string]protocol.Task{"app:start": {}}, "app:start")
	g.Tasks["app:start"].Fixing = true
	if got := AttachedTask(g, true, false); got != "" {
		t.Errorf("a lint fix: got %q", got)
	}
	if got := AttachedTask(nil, true, false); got != "" {
		t.Errorf("no graph: got %q", got)
	}
}

func TestWireGivesTheTerminalOrTheWriters(t *testing.T) {
	cmd := exec.Command("x")
	wire(cmd, Spec{Terminal: true, Out: &bytes.Buffer{}})
	if cmd.Stdin != os.Stdin || cmd.Stdout != os.Stdout || cmd.Stderr != os.Stderr {
		t.Fatal("an attached command uses smith's own streams")
	}
	if cmd.SysProcAttr != nil {
		t.Fatal("an attached command stays in the terminal's process group")
	}
	var out bytes.Buffer
	cmd = exec.Command("x")
	wire(cmd, Spec{Out: &out})
	if cmd.Stdin != nil || cmd.Stdout != &out || cmd.Stderr != &out {
		t.Fatal("a command not attached writes to Out and reads nothing")
	}
	if cmd.SysProcAttr == nil {
		t.Fatal("a command not attached gets a process group of its own")
	}
}
