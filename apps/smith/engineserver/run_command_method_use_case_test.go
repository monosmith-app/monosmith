package engineserver

import (
	"context"
	"fmt"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/runevents"
)

// recording is a server whose smith commands are only noted, never run.
func recording(ran *[][]string, stdout string, code int) *Server {
	return &Server{Version: "1", Execute: func(_ context.Context, _ string, args []string, onLine func(runevents.Stream, string)) ([]byte, int, error) {
		*ran = append(*ran, args)
		if onLine != nil {
			onLine(runevents.Stdout, "step")
		}
		return []byte(stdout), code, nil
	}}
}

func TestReleaseIsAStreamedRunThatDoesNotAskAgain(t *testing.T) {
	var ran [][]string
	c := start(t, recording(&ran, "", 0))
	result := c.call("release", map[string]any{"run": "rel-1", "noPush": true})["result"].(map[string]any)
	if result["ok"] != true || result["run"] != "rel-1" {
		t.Fatalf("release answered %v", result)
	}
	want := []string{"run.started", "task.queued release", "task.started release", "task.output release", "task.finished release passed", "run.finished"}
	if fmt.Sprint(c.kinds()) != fmt.Sprint(want) {
		t.Fatalf("events\n got %q\nwant %q", c.kinds(), want)
	}
	if first := c.events[0]; first.Run != "rel-1" || first.Command != "smith release --yes --no-push" {
		t.Fatalf("run.started %+v", first)
	}
	c.call("release", nil)
	if fmt.Sprint(ran) != "[[release --yes --no-push] [release --yes]]" {
		t.Fatalf("commands run %v", ran)
	}
}

func TestAFailedReleaseEndsTheRunFailedWithItsExitCode(t *testing.T) {
	var ran [][]string
	c := start(t, recording(&ran, "", 1))
	result := c.call("release", nil)["result"].(map[string]any)
	if result["ok"] != false || result["exitCode"] != float64(1) {
		t.Fatalf("release answered %v", result)
	}
	if last := c.events[len(c.events)-1]; last.Event != runevents.RunFinished || last.OK == nil || *last.OK {
		t.Fatalf("run.finished %+v", last)
	}
}

func TestConfigSetAndUnsetAreRunsOverTheConfigCommand(t *testing.T) {
	var ran [][]string
	c := start(t, recording(&ran, "", 0))
	c.call("config.set", map[string]any{"setting": "ci.devcontainer", "value": "false"})
	c.call("config.set", map[string]any{"setting": "ci.providers", "values": []string{"github", "azure"}})
	c.call("config.unset", map[string]any{"setting": "release.branch"})
	if fmt.Sprint(ran) != "[[config set -- ci.devcontainer false] [config set -- ci.providers github azure] [config unset release.branch]]" {
		t.Fatalf("commands run %v", ran)
	}
	for _, params := range []map[string]any{{"value": "x"}, {"setting": "--force", "value": "x"}, {"setting": "ci.devcontainer"}} {
		if reply := c.call("config.set", params); reply["error"].(map[string]any)["code"].(float64) != codeInvalidParams {
			t.Errorf("config.set %v answered %v", params, reply)
		}
	}
	if len(ran) != 3 {
		t.Fatalf("a refused request ran %v", ran)
	}
}
