package engineserver

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/mcpserver"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/runevents"
	"github.com/monosmith-app/monosmith/apps/smith/taskgraph"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/libs/actions/catalog"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TestMain lets the test binary act as a task command: with SMITH_HELPER set it
// prints a line, and for "slow" it waits two seconds before writing out/done.
func TestMain(m *testing.M) {
	if os.Getenv("SMITH_HELPER") != "" {
		arg := os.Args[len(os.Args)-1]
		fmt.Println("helper ran with", arg)
		switch arg {
		case "fail":
			os.Exit(3)
		case "slow":
			time.Sleep(2 * time.Second)
			_ = os.MkdirAll("out", 0o755)
			_ = os.WriteFile(filepath.Join("out", "done"), []byte("done"), 0o644)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// client is a JSON-RPC client of a Server over pipes, as an editor would be.
type client struct {
	t       *testing.T
	in      io.WriteCloser
	mu      sync.Mutex
	events  []runevents.Event
	changed chan struct{}
	replies map[int]chan map[string]any
	done    chan error
	nextID  int
}

func start(t *testing.T, s *Server) *client {
	t.Helper()
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	c := &client{t: t, in: inW, changed: make(chan struct{}, 1024), replies: map[int]chan map[string]any{}, done: make(chan error, 1)}
	go func() {
		err := s.Serve(context.Background(), inR, outW)
		_ = outW.Close()
		c.done <- err
	}()
	go func() {
		sc := bufio.NewScanner(outR)
		sc.Buffer(make([]byte, 1<<20), 16<<20)
		for sc.Scan() {
			var msg map[string]any
			if err := json.Unmarshal(sc.Bytes(), &msg); err != nil {
				t.Errorf("the server wrote %q, which is not JSON", sc.Text())
				continue
			}
			c.mu.Lock()
			if msg["method"] == EventNotification {
				var e runevents.Event
				raw, _ := json.Marshal(msg["params"])
				_ = json.Unmarshal(raw, &e)
				c.events = append(c.events, e)
				c.changed <- struct{}{}
			} else if id, ok := msg["id"].(float64); ok {
				c.replies[int(id)] <- msg
			}
			c.mu.Unlock()
		}
	}()
	t.Cleanup(func() { _ = inW.Close() })
	return c
}

// send writes a request and returns the channel its answer arrives on.
func (c *client) send(method string, params any) chan map[string]any {
	c.t.Helper()
	c.mu.Lock()
	c.nextID++
	id := c.nextID
	ch := make(chan map[string]any, 1)
	c.replies[id] = ch
	c.mu.Unlock()
	line, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params})
	if _, err := c.in.Write(append(line, '\n')); err != nil {
		c.t.Fatal(err)
	}
	return ch
}

func (c *client) call(method string, params any) map[string]any {
	c.t.Helper()
	return c.await(c.send(method, params))
}

func (c *client) await(ch chan map[string]any) map[string]any {
	c.t.Helper()
	select {
	case msg := <-ch:
		return msg
	case <-time.After(30 * time.Second):
		c.t.Fatal("no answer")
		return nil
	}
}

func (c *client) waitFor(match func(runevents.Event) bool) {
	c.t.Helper()
	deadline := time.After(30 * time.Second)
	for {
		c.mu.Lock()
		for _, e := range c.events {
			if match(e) {
				c.mu.Unlock()
				return
			}
		}
		c.mu.Unlock()
		select {
		case <-c.changed:
		case <-deadline:
			c.t.Fatal("the event never arrived")
		}
	}
}

func (c *client) kinds() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []string
	for _, e := range c.events {
		out = append(out, strings.TrimSpace(fmt.Sprintf("%s %s %s", e.Event, e.Task, e.Status)))
	}
	return out
}

func first(s *Server, _ string) *Server { return s }

// workspace is a temporary workspace with a project whose build task is the
// test binary acting as the helper, and a server whose runs use it.
func workspace(t *testing.T, arg string) (*Server, string) {
	t.Helper()
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "apps", "web"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "apps", "web", "src.txt"), []byte("v1"), 0o644)
	self, _ := os.Executable()
	web := &projectgraph.Project{Project: protocol.Project{Name: "web", Root: "apps/web", Tasks: map[string]protocol.Task{
		"build": {Command: []string{self, "-test.run=^$", arg}, Env: map[string]string{"SMITH_HELPER": "1"}},
	}}, Plugin: "test"}
	pg, err := projectgraph.New([]*projectgraph.Project{web})
	if err != nil {
		t.Fatal(err)
	}
	return &Server{Version: "9.9.9", RunTasks: func(ctx context.Context, req taskrun.Request) (taskrun.Summary, error) {
		g, err := taskgraph.Build(pg, []string{"web"}, req.Tasks)
		if err != nil {
			return taskrun.Summary{}, err
		}
		return taskrun.Run(ctx, g, taskrun.Options{
			Root: root, Files: []string{"apps/web/src.txt"}, Projects: pg, Out: io.Discard, ToolVersion: "test",
			Events: runevents.NewRunStream(req.RunID, req.Events), Command: req.Command(),
		}), nil
	}}, root
}

func TestARunStreamsOrderedEventsAndAnswersWithTheResult(t *testing.T) {
	c := start(t, first(workspace(t, "hello")))
	init := c.call("initialize", nil)["result"].(map[string]any)
	if init["name"] != "smith" || init["version"] != "9.9.9" || int(init["protocolVersion"].(float64)) != ProtocolVersion {
		t.Fatalf("initialize %v", init)
	}

	reply := c.call("run", RunParams{Run: "r1", Tasks: []string{"build"}})
	result, ok := reply["result"].(map[string]any)
	if !ok || result["ok"] != true || result["run"] != "r1" {
		t.Fatalf("run answered %v", reply)
	}
	if got := result["summary"].(map[string]any)["results"].([]any); len(got) != 1 || got[0].(map[string]any)["status"] != "success" {
		t.Fatalf("summary %v", result["summary"])
	}
	want := []string{"run.started", "task.queued web:build", "task.started web:build", "task.output web:build", "task.finished web:build passed", "run.finished"}
	if fmt.Sprint(c.kinds()) != fmt.Sprint(want) {
		t.Fatalf("events\n got %q\nwant %q", c.kinds(), want)
	}
	for i, e := range c.events {
		if e.Run != "r1" || e.Seq != i+1 {
			t.Fatalf("event %d: %+v", i, e)
		}
	}
	if c.events[3].Line != "helper ran with hello" {
		t.Fatalf("output %+v", c.events[3])
	}
	if last := c.events[5]; last.OK == nil || !*last.OK || last.Summary.Passed != 1 {
		t.Fatalf("run.finished %+v", last)
	}
}

func TestAFailedRunIsAnAnswerNotAnError(t *testing.T) {
	c := start(t, first(workspace(t, "fail")))
	result := c.call("run", RunParams{Tasks: []string{"build"}})["result"].(map[string]any)
	if result["ok"] != false {
		t.Fatalf("run answered %v", result)
	}
	c.waitFor(func(e runevents.Event) bool {
		return e.Event == runevents.TaskFinished && e.Status == runevents.Failed && e.Reason == "exited with code 3"
	})
}

func TestCancellingARunStopsItsProcesses(t *testing.T) {
	s, root := workspace(t, "slow")
	c := start(t, s)
	answer := c.send("run", RunParams{Run: "slow-run", Tasks: []string{"build"}})
	c.waitFor(func(e runevents.Event) bool { return e.Event == runevents.TaskOutput })

	cancelled := c.call("cancel", CancelParams{Run: "slow-run"})["result"].(map[string]any)
	if cancelled["cancelled"] != true {
		t.Fatalf("cancel answered %v", cancelled)
	}
	result := c.await(answer)["result"].(map[string]any)
	if result["ok"] != false {
		t.Fatalf("a cancelled run is not ok: %v", result)
	}
	c.waitFor(func(e runevents.Event) bool { return e.Event == runevents.RunFinished })
	c.waitFor(func(e runevents.Event) bool {
		return e.Event == runevents.TaskFinished && e.Status == runevents.Failed && e.Reason == "cancelled"
	})
	if again := c.call("cancel", CancelParams{Run: "slow-run"})["result"].(map[string]any); again["cancelled"] != false {
		t.Fatalf("a finished run cannot be cancelled: %v", again)
	}

	// The task would have written out/done after two seconds if it had lived.
	time.Sleep(3 * time.Second)
	if _, err := os.Stat(filepath.Join(root, "apps", "web", "out", "done")); err == nil {
		t.Fatal("the process outlived the cancel")
	}
}

func TestQueriesAnswerWithTheJSONOfTheirCommand(t *testing.T) {
	var ran [][]string
	s := &Server{Version: "1", Execute: func(_ context.Context, _ string, args []string, _ func(runevents.Stream, string)) ([]byte, int, error) {
		ran = append(ran, args)
		return []byte(`{"projects":[{"name":"web"}]}` + "\n"), 1, nil
	}}
	c := start(t, s)
	reply := c.call("projects", nil)
	if got := reply["result"].(map[string]any)["projects"].([]any)[0].(map[string]any)["name"]; got != "web" {
		t.Fatalf("projects answered %v", reply)
	}
	c.call("explain", map[string]any{"target": "web:build"})
	c.call("release.preview", nil)
	if fmt.Sprint(ran) != "[[projects --json] [explain web:build --json] [release --dry-run --json]]" {
		t.Fatalf("commands run %v", ran)
	}
	if missing := c.call("no.such.method", nil); missing["error"].(map[string]any)["code"].(float64) != codeMethodNotFound {
		t.Fatalf("unknown method answered %v", missing)
	}
}

func TestChangingAKindIsAPreviewQueryAndACommand(t *testing.T) {
	var ran [][]string
	s := &Server{Version: "1", Execute: func(_ context.Context, _ string, args []string, _ func(runevents.Stream, string)) ([]byte, int, error) {
		ran = append(ran, args)
		return []byte(`{"project":"core"}` + "\n"), 0, nil
	}}
	c := start(t, s)
	if preview := c.call("kind.preview", map[string]any{"project": "core", "kind": "package"}); preview["result"].(map[string]any)["project"] != "core" {
		t.Fatalf("kind.preview answered %v", preview)
	}
	if result := c.call("kind", map[string]any{"project": "core", "kind": "lib"})["result"].(map[string]any); result["ok"] != true {
		t.Fatalf("kind answered %v", result)
	}
	if fmt.Sprint(ran) != "[[kind core package --dry-run --json] [kind core lib --json]]" {
		t.Fatalf("commands run %v", ran)
	}
	if bad := c.call("kind", map[string]any{"project": "core", "kind": "--force"}); bad["error"] == nil {
		t.Fatalf("an option smuggled in as the kind must be refused: %v", bad)
	}
}

func TestACommandIsARunWithItsOutputAsEvents(t *testing.T) {
	s := &Server{Version: "1", Execute: func(_ context.Context, _ string, args []string, onLine func(runevents.Stream, string)) ([]byte, int, error) {
		onLine(runevents.Stdout, "added zod")
		onLine(runevents.Stderr, "a warning")
		return nil, 0, nil
	}}
	c := start(t, s)
	result := c.call("install", map[string]any{"package": "zod", "projects": []string{"web"}})["result"].(map[string]any)
	if result["ok"] != true {
		t.Fatalf("install answered %v", result)
	}
	want := []string{"run.started", "task.queued install", "task.started install", "task.output install", "task.output install", "task.finished install passed", "run.finished"}
	if fmt.Sprint(c.kinds()) != fmt.Sprint(want) {
		t.Fatalf("events\n got %q\nwant %q", c.kinds(), want)
	}
	if first := c.events[0]; !strings.HasPrefix(first.Command, "smith install") {
		t.Fatalf("run.started %+v", first)
	}
	if out := c.events[3]; out.Stream != runevents.Stdout || out.Line != "added zod" {
		t.Fatalf("task.output %+v", out)
	}
}

func TestShutdownAnswersAndEndsTheServer(t *testing.T) {
	c := start(t, &Server{Version: "1"})
	c.call("shutdown", nil)
	select {
	case err := <-c.done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the server did not stop")
	}
}

// served is the id of every action a method of the engine runs.
func served(t *testing.T) map[string]bool {
	t.Helper()
	out := map[string]bool{"run": true}
	toolAction := func(method, tool string) {
		for _, tl := range mcpserver.Tools {
			if tl.Name == tool {
				out[tl.Action] = true
				return
			}
		}
		t.Errorf("method %s maps to the tool %q, which does not exist", method, tool)
	}
	for method, tool := range queryMethods {
		toolAction(method, tool)
	}
	for method, tool := range commandMethods {
		toolAction(method, tool)
	}
	for method := range serverCommands {
		if _, ok := serverActions[method]; !ok {
			t.Errorf("method %s has no action in serverActions", method)
		}
	}
	for method := range serverMethods {
		if _, ok := serverActions[method]; !ok {
			t.Errorf("method %s has no action in serverActions", method)
		}
	}
	for method, id := range serverActions {
		if _, ok := catalog.ByID(id); !ok {
			t.Errorf("method %s runs the action %q, which does not exist", method, id)
		}
		out[id] = true
	}
	return out
}

func TestEveryActionHasAMethodOrIsExcused(t *testing.T) {
	shown := served(t)
	for _, a := range catalog.All() {
		switch {
		case shown[a.ID()] && a.NotInServe != "":
			t.Errorf("action %q has an engine method and is also excused (%q): drop NotInServe", a.ID(), a.NotInServe)
		case !shown[a.ID()] && a.NotInServe == "":
			t.Errorf("action %q has no engine method: add one, or say why not in NotInServe (libs/actions/catalog/catalog_config.go)", a.ID())
		}
	}
}

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

func TestConfigListAnswersWithTheJSONOfSmithConfigList(t *testing.T) {
	var ran [][]string
	c := start(t, recording(&ran, `[{"path":"ci.devcontainer","set":false}]`, 0))
	rows := c.call("config.list", nil)["result"].([]any)
	if rows[0].(map[string]any)["path"] != "ci.devcontainer" || fmt.Sprint(ran) != "[[config list --json]]" {
		t.Fatalf("config.list answered %v after %v", rows, ran)
	}
}

func TestConfigGetReadsASettingAndSaysWhetherItIsSet(t *testing.T) {
	t.Setenv("SMITH_WORKSPACE_ROOT", "") // smith sets it for the tasks it runs
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "smith.toml"), []byte("schema = 1\n\n[workspace]\nname = \"acme\"\n\n[ci]\ndevcontainer = false\n"), 0o644)
	c := start(t, &Server{Version: "1", Dir: root})

	set := c.call("config.get", map[string]any{"setting": "ci.devcontainer"})["result"].(map[string]any)
	if set["set"] != true || set["value"] != "false" || set["type"] != "boolean" || set["description"] == "" {
		t.Fatalf("config.get answered %v", set)
	}
	unset := c.call("config.get", map[string]any{"setting": "release.branch"})["result"].(map[string]any)
	if unset["set"] != false || unset["setting"] != "release.branch" {
		t.Fatalf("config.get answered %v", unset)
	}
	if reply := c.call("config.get", map[string]any{"setting": "stacks.web"}); reply["error"].(map[string]any)["code"].(float64) != codeInvalidParams {
		t.Fatalf("an unknown setting answered %v", reply)
	}
}

const teamRecipeFile = `name = "team"
version = "1.2.0"

[choices.language]
ask = "Language"
options = ["python", "csharp"]

[use]
structure = "vertical-slices"

[when.language.python.use]
language = "python"

[when.language.csharp.use]
language = "dotnet"

[when.language.csharp.settings.dotnet]
target_framework = "net10.0"
`

func recipeWorkspace(t *testing.T) *client {
	t.Helper()
	t.Setenv("SMITH_WORKSPACE_ROOT", "") // smith sets it for the tasks it runs
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "smith.toml"), []byte("schema = 1\n\n[workspace]\nname = \"acme\"\n"), 0o644)
	_ = os.MkdirAll(filepath.Join(root, "recipes"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "recipes", "team.recipe.toml"), []byte(teamRecipeFile), 0o644)
	return start(t, &Server{Version: "1", Dir: root})
}

func TestRecipeListFindsTheWorkspacesRecipes(t *testing.T) {
	c := recipeWorkspace(t)
	recipes := c.call("recipe.list", nil)["result"].(map[string]any)["recipes"].([]any)
	if len(recipes) != 1 {
		t.Fatalf("recipe.list answered %v", recipes)
	}
	if r := recipes[0].(map[string]any); r["path"] != "recipes/team.recipe.toml" || r["name"] != "team" || r["version"] != "1.2.0" {
		t.Fatalf("recipe.list answered %v", r)
	}
}

func TestRecipeDescribeShowsWhatARecipeOffersBeforeAnyAnswer(t *testing.T) {
	c := recipeWorkspace(t)
	d := c.call("recipe.describe", map[string]any{"recipe": "recipes/team.recipe.toml"})["result"].(map[string]any)
	if d["name"] != "team" || len(d["choices"].([]any)) != 1 || fmt.Sprint(d["configures"]) != "[dotnet]" {
		t.Fatalf("recipe.describe answered %v", d)
	}
	var language map[string]any
	for _, u := range d["use"].([]any) {
		if u.(map[string]any)["key"] == "language" {
			language = u.(map[string]any)
		}
	}
	if fmt.Sprint(language["alternatives"]) != "map[language=csharp:[dotnet] language=python:[python]]" {
		t.Fatalf("language role %v", language)
	}
}

func TestRecipeRolesSaysWhatTheRecipePutsInEachRoleForTheAnswers(t *testing.T) {
	c := recipeWorkspace(t)
	r := c.call("recipe.roles", map[string]any{"recipe": "recipes/team.recipe.toml", "choices": map[string]string{"language": "csharp"}})["result"].(map[string]any)
	got := map[string]string{}
	for _, role := range r["roles"].([]any) {
		m := role.(map[string]any)
		got[m["key"].(string)] = fmt.Sprint(m["plugins"])
	}
	if fmt.Sprint(got) != "map[language:[dotnet] structure:[vertical-slices]]" {
		t.Fatalf("roles %v", got)
	}
	if fmt.Sprint(r["choices"].([]any)[0].(map[string]any)["answer"]) != "csharp" {
		t.Fatalf("choices %v", r["choices"])
	}
	// A choice without a default has to be answered.
	missing := c.call("recipe.roles", map[string]any{"recipe": "recipes/team.recipe.toml"})
	if msg := missing["error"].(map[string]any)["message"].(string); !strings.Contains(msg, "language") {
		t.Fatalf("an unanswered choice answered %v", missing)
	}
}

func TestARecipeOutsideTheWorkspaceIsRefused(t *testing.T) {
	c := recipeWorkspace(t)
	for _, ref := range []string{"../team.recipe.toml", "--all", "http://example.com/x.recipe.toml"} {
		reply := c.call("recipe.describe", map[string]any{"recipe": ref})
		if reply["error"] == nil {
			t.Errorf("%s answered %v", ref, reply)
		}
	}
}
