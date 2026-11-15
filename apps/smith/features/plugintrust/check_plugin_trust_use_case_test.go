package plugintrust

import (
	"path/filepath"
	"testing"
)

func TestCheckFollowsTheFileAndTheCommand(t *testing.T) {
	dir, ws := t.TempDir(), t.TempDir()
	store := &Store{Path: filepath.Join(t.TempDir(), "trust.json")}
	plugin := writeFile(t, dir, "plugin", "v1")
	req := Request{Workspace: ws, Plugin: "w", Command: []string{plugin}}

	d, err := Check(store, req)
	if err != nil || d.Trusted || d.Reason != ReasonNew {
		t.Fatalf("first use: %+v %v", d, err)
	}
	if err := store.Trust(ws, "w", d.Fingerprint); err != nil {
		t.Fatal(err)
	}
	if d, _ = Check(store, req); !d.Trusted || d.Via != "recorded" {
		t.Fatalf("after trusting: %+v", d)
	}

	// Another workspace does not inherit the decision.
	other := req
	other.Workspace = t.TempDir()
	if d, _ = Check(store, other); d.Trusted {
		t.Fatalf("another workspace: %+v", d)
	}

	// A different command line is untrusted again.
	withArg := req
	withArg.Command = []string{plugin, "--serve"}
	if d, _ = Check(store, withArg); d.Trusted || d.Reason != ReasonCommandChanged {
		t.Fatalf("changed command: %+v", d)
	}

	// So is the same command whose file changed.
	writeFile(t, dir, "plugin", "v2")
	if d, _ = Check(store, req); d.Trusted || d.Reason != ReasonFileChanged {
		t.Fatalf("changed file: %+v", d)
	}
}

func TestCheckInCIUsesThePinAndIgnoresTheStore(t *testing.T) {
	dir, ws := t.TempDir(), t.TempDir()
	store := &Store{Path: filepath.Join(t.TempDir(), "trust.json")}
	plugin := writeFile(t, dir, "plugin", "v1")
	req := Request{Workspace: ws, Plugin: "w", Command: []string{plugin}, InCI: true}
	fpNow, _ := Measure(req.Command)
	_ = store.Trust(ws, "w", fpNow)

	if d, _ := Check(store, req); d.Trusted || d.Reason != ReasonNotPinned {
		t.Fatalf("unpinned in CI: %+v", d)
	}
	req.Pin = digest("v1")
	if d, _ := Check(store, req); !d.Trusted || d.Via != "pinned" {
		t.Fatalf("pinned in CI: %+v", d)
	}
	req.Pin = digest("other")
	if d, _ := Check(store, req); d.Trusted || d.Reason != ReasonPinMismatch {
		t.Fatalf("wrong pin in CI: %+v", d)
	}
}

func TestCheckFailsWhenTheProgramIsMissing(t *testing.T) {
	store := &Store{Path: filepath.Join(t.TempDir(), "trust.json")}
	if _, err := Check(store, Request{Workspace: "/w", Plugin: "w", Command: []string{filepath.Join(t.TempDir(), "missing")}}); err == nil {
		t.Fatal("no error")
	}
}
