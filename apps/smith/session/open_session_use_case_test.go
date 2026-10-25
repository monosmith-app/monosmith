package session

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/pluginadmit"
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
	"github.com/monosmith-app/monosmith/apps/smith/remotecache"
	"github.com/monosmith-app/monosmith/apps/smith/taskcache"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// TestMain lets the test binary stand in for smith: the plugin host starts
// built-in plugins as `<self> plugin serve <name>`.
func TestMain(m *testing.M) {
	if len(os.Args) == 4 && os.Args[1] == "plugin" && os.Args[2] == "serve" {
		ctor, ok := pluginhost.Builtin(os.Args[3])
		if !ok {
			os.Exit(2)
		}
		if err := serve.Serve(ctor("test"), os.Stdin, os.Stdout); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// widgetWorkspace is a real git workspace that declares the Go example plugin
// (built from docs/examples/plugins/go) and holds one widget project.
func widgetWorkspace(t *testing.T) (root, plugin string) {
	t.Helper()
	examples, err := filepath.Abs(filepath.Join("..", "..", "..", "docs", "examples", "plugins", "go"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(examples); err != nil {
		t.Skip("the examples are not part of this checkout")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not installed")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	for _, v := range []string{"CI", "GITHUB_ACTIONS", "TF_BUILD", "SMITH_WORKSPACE_ROOT"} {
		t.Setenv(v, "")
	}
	t.Setenv("SMITH_HOME", t.TempDir())
	pluginadmit.SetPrompter(nil)

	plugin = filepath.Join(t.TempDir(), "smith-plugin-widget")
	if runtime.GOOS == "windows" {
		plugin += ".exe"
	}
	build := exec.Command("go", "build", "-o", plugin, ".")
	build.Dir = examples
	build.Env = append(os.Environ(), "GOWORK=off")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building the example plugin: %v\n%s", err, out)
	}

	root = t.TempDir()
	toml := "[workspace]\nname = \"wd\"\n\n[plugins.external.widget]\ncommand = ['" + plugin + "']\n"
	for name, body := range map[string]string{workspace.FileName: toml, "apps/hello/widget.json": `{"name":"hello"}`} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	return root, plugin
}

func TestAnExternalPluginIsRefusedUntilItIsTrustedAndAgainWhenItChanges(t *testing.T) {
	root, plugin := widgetWorkspace(t)
	open := func() (*Session, error) {
		return Open(Options{Dir: root, Exact: true, Version: "test", Log: io.Discard})
	}

	// Untrusted: refused, with the command that fixes it, and nothing runs.
	_, err := open()
	if !errors.Is(err, pluginadmit.ErrUntrusted) || !strings.Contains(err.Error(), "smith plugin trust widget") {
		t.Fatalf("an untrusted plugin was not refused: %v", err)
	}

	// Trusted: the plugin starts and finds the project.
	store, err := plugintrust.OpenStore()
	if err != nil {
		t.Fatal(err)
	}
	fp, err := plugintrust.Measure([]string{plugin})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Trust(root, "widget", fp); err != nil {
		t.Fatal(err)
	}
	s, err := open()
	if err != nil {
		t.Fatalf("a trusted plugin did not start: %v", err)
	}
	found := false
	for _, p := range s.Graph.Projects {
		found = found || p.Name == "hello"
	}
	s.Close()
	if !found {
		t.Fatal("the external plugin did not discover its project")
	}

	// Changed: trusted no more.
	f, err := os.OpenFile(plugin, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString("tampered")
	_ = f.Close()
	if _, err := open(); !errors.Is(err, pluginadmit.ErrUntrusted) || !strings.Contains(err.Error(), "changed since you trusted it") {
		t.Fatalf("a changed plugin was not refused: %v", err)
	}
}

func TestInCIAnExternalPluginStartsOnlyWhenPinned(t *testing.T) {
	root, plugin := widgetWorkspace(t)
	t.Setenv("CI", "true")
	open := func() (*Session, error) {
		return Open(Options{Dir: root, Exact: true, Version: "test", Log: io.Discard})
	}
	if _, err := open(); !errors.Is(err, pluginadmit.ErrUntrusted) || !strings.Contains(err.Error(), "not pinned") {
		t.Fatalf("an unpinned plugin started in CI: %v", err)
	}
	fp, err := plugintrust.Measure([]string{plugin})
	if err != nil {
		t.Fatal(err)
	}
	toml := "[workspace]\nname = \"wd\"\n\n[plugins.external.widget]\ncommand = ['" + plugin + "']\nsha256 = \"" + plugintrust.PinnedDigest(fp) + "\"\n"
	if err := os.WriteFile(filepath.Join(root, workspace.FileName), []byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := open()
	if err != nil {
		t.Fatalf("a pinned plugin did not start in CI: %v", err)
	}
	s.Close()
}

func remoteSession(t *testing.T) *Session {
	t.Helper()
	cfg := workspace.Config{}
	cfg.Cache.Remote = "https://cache.example.invalid/smith"
	cfg.Cache.Dir = t.TempDir()
	return &Session{Root: t.TempDir(), Config: cfg}
}

func clearCacheEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range append([]string{"SMITH_CACHE_REMOTE", "SMITH_CACHE_READ_ONLY", "GITHUB_EVENT_NAME", "BUILD_REASON", "CI_PIPELINE_SOURCE"}, cacheEnvironment...) {
		t.Setenv(name, "")
	}
}

func TestAnEmptyCacheKeyLeavesTheRunWorkingWithTheRemoteOff(t *testing.T) {
	// GitHub passes a fork's pull request an empty secret.
	clearCacheEnvironment(t)
	storage, err := remoteSession(t).Storage()
	if err != nil {
		t.Fatal(err)
	}
	if _, local := storage.(*taskcache.LocalStorage); !local {
		t.Errorf("storage is %T, want the local store alone", storage)
	}
}

func TestAnUnexpandedAzureMacroLeavesTheRemoteOff(t *testing.T) {
	// Azure Pipelines hands a fork's build the variable's own name.
	clearCacheEnvironment(t)
	t.Setenv("SMITH_CACHE_KEY", "$(SMITH_CACHE_KEY)")
	t.Setenv("SMITH_CACHE_TOKEN", "$(SMITH_CACHE_TOKEN)")
	storage, err := remoteSession(t).Storage()
	if err != nil {
		t.Fatal(err)
	}
	if _, local := storage.(*taskcache.LocalStorage); !local {
		t.Errorf("storage is %T, want the local store alone", storage)
	}
}

func TestPullRequestsReadTheRemoteAndNeverWriteIt(t *testing.T) {
	for name, env := range map[string]map[string]string{
		"github": {"GITHUB_EVENT_NAME": "pull_request"},
		"azure":  {"BUILD_REASON": "PullRequest"},
		"gitlab": {"CI_PIPELINE_SOURCE": "merge_request_event"},
	} {
		clearCacheEnvironment(t)
		t.Setenv("SMITH_CACHE_KEY", "a-shared-secret")
		for key, value := range env {
			t.Setenv(key, value)
		}
		storage, err := remoteSession(t).Storage()
		if err != nil {
			t.Fatal(err)
		}
		tiered, ok := storage.(*remotecache.Tiered)
		if !ok || !tiered.ReadOnly {
			t.Errorf("%s: storage is %T (read only: %v), want a read-only remote", name, storage, ok && tiered.ReadOnly)
		}
	}
}

func TestThePushToTheReleaseBranchWritesTheRemote(t *testing.T) {
	clearCacheEnvironment(t)
	t.Setenv("SMITH_CACHE_KEY", "a-shared-secret")
	t.Setenv("GITHUB_EVENT_NAME", "push")
	storage, err := remoteSession(t).Storage()
	if err != nil {
		t.Fatal(err)
	}
	if tiered, ok := storage.(*remotecache.Tiered); !ok || tiered.ReadOnly {
		t.Errorf("storage is %T, want a writable remote", storage)
	}
}

// cyclicWorkspace is a git workspace whose two Node projects import each other.
func cyclicWorkspace(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	for _, v := range []string{"CI", "GITHUB_ACTIONS", "TF_BUILD", "SMITH_WORKSPACE_ROOT"} {
		t.Setenv(v, "")
	}
	t.Setenv("SMITH_HOME", t.TempDir())
	root := t.TempDir()
	files := map[string]string{
		workspace.FileName:      "[workspace]\nname = \"cyc\"\n",
		"package.json":          `{"name":"root","private":true,"workspaces":["libs/*"]}`,
		"libs/ui/package.json":  `{"name":"ui","version":"1.0.0","dependencies":{"web":"*"}}`,
		"libs/web/package.json": `{"name":"web","version":"1.0.0","dependencies":{"ui":"*"}}`,
	}
	for name, body := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	return root
}

func TestACycleStopsEveryCommandThatNeedsAnOrder(t *testing.T) {
	root := cyclicWorkspace(t)
	s, err := Open(Options{Dir: root, Exact: true, Version: "test", Log: io.Discard})
	if err == nil {
		s.Close()
		t.Fatal("a workspace with an import cycle opened")
	}
	if !strings.Contains(err.Error(), "projects import each other in a cycle: ui → web → ui") {
		t.Fatalf("the error does not name the cycle: %v", err)
	}
}

func TestACycleLeavesTheWorkspaceOpenForTheCommandsThatReportIt(t *testing.T) {
	root := cyclicWorkspace(t)
	s, err := Open(Options{Dir: root, Exact: true, Version: "test", Log: io.Discard, AllowCycles: true})
	if err != nil {
		t.Fatalf("doctor and graph could not open a cyclic workspace: %v", err)
	}
	defer s.Close()
	cycles := s.Graph.Cycles()
	if len(cycles) != 1 || strings.Join(cycles[0], " ") != "ui web ui" {
		t.Fatalf("cycles %v", cycles)
	}
	if _, err := s.Graph.Order(); err == nil {
		t.Fatal("Order must still refuse a cycle")
	}
	// A refresh keeps the choice.
	if err := s.Refresh(); err != nil {
		t.Fatalf("refresh of a cyclic workspace: %v", err)
	}
}
