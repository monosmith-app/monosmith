package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/conformance"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

func connect(t *testing.T, workspace string, delay time.Duration) *host.Client {
	t.Helper()
	hostR, pluginW := io.Pipe()
	pluginR, hostW := io.Pipe()
	go func() { _ = serve.Serve(&scan{delay: delay}, pluginR, pluginW); _ = pluginW.Close() }()
	c, err := host.Connect(hostR, hostW, host.Options{WorkspaceRoot: workspace})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

// project writes files under libs/core and returns the project to scan.
func project(t *testing.T, workspace string, files map[string]int) protocol.Project {
	t.Helper()
	for name, size := range files {
		path := filepath.Join(workspace, "libs", "core", filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return protocol.Project{Name: "core", Root: "libs/core", Language: "widget", Kind: protocol.ProjectLib}
}

func TestConforms(t *testing.T) {
	report := conformance.Check(connect(t, t.TempDir(), 0), conformance.Options{})
	if !report.OK() {
		t.Fatal(report.Failures())
	}
}

func TestAdvertisesCancelAndAConcurrencyOfFour(t *testing.T) {
	c := connect(t, t.TempDir(), 0)
	if !c.Info.Capabilities.Cancel || c.Info.Concurrency != 4 {
		t.Fatalf("cancel %v, concurrency %d", c.Info.Capabilities.Cancel, c.Info.Concurrency)
	}
}

func TestDoctorWarnsAboutABigFile(t *testing.T) {
	ws := t.TempDir()
	core := project(t, ws, map[string]int{"src/small.txt": 10, "data/huge.bin": bigFile + 1})
	var res protocol.DoctorResult
	if err := connect(t, ws, 0).Call(protocol.MethodDoctor, protocol.DoctorParams{Projects: []protocol.Project{core}}, &res, 0); err != nil {
		t.Fatal(err)
	}
	f := res.Findings[0]
	if f.OK || !f.Warning || f.Detail != "data/huge.bin" {
		t.Fatalf("%+v", f)
	}
}

func TestACancelledScanStopsMidwayAndTheAnswerIsTheCancelledError(t *testing.T) {
	ws := t.TempDir()
	files := map[string]int{}
	for _, name := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		files[name+".txt"] = 1
	}
	core := project(t, ws, files)
	c := connect(t, ws, 100*time.Millisecond) // eight files: the scan would take 800ms
	ctx, cancel := context.WithCancel(context.Background())
	failed := make(chan error, 1)
	began := time.Now()
	go func() {
		failed <- c.CallContext(ctx, protocol.MethodDoctor, protocol.DoctorParams{Projects: []protocol.Project{core}}, nil, 0)
	}()
	time.Sleep(150 * time.Millisecond)
	cancel()
	err := <-failed
	var rpcErr *protocol.ResponseError
	if !errors.Is(err, context.Canceled) || !errors.As(err, &rpcErr) || rpcErr.Code != protocol.CodeRequestCancelled {
		t.Fatalf("want a cancelled scan, got %v", err)
	}
	if took := time.Since(began); took > 600*time.Millisecond {
		t.Fatalf("the scan went on for %s after being cancelled", took)
	}
}

func TestFourScansRunAtTheSameTime(t *testing.T) {
	ws := t.TempDir()
	core := project(t, ws, map[string]int{"a.txt": 1, "b.txt": 1, "c.txt": 1})
	c := connect(t, ws, 100*time.Millisecond)
	began := time.Now()
	done := make(chan error, 4)
	for range 4 {
		go func() {
			done <- c.Call(protocol.MethodDoctor, protocol.DoctorParams{Projects: []protocol.Project{core}}, nil, 0)
		}()
	}
	for range 4 {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	// Three files at 100ms each is 300ms a scan: one after the other they take 1.2s.
	if took := time.Since(began); took > 900*time.Millisecond || c.PeakInFlight() != 4 {
		t.Fatalf("four scans took %s with %d in flight", took, c.PeakInFlight())
	}
}
