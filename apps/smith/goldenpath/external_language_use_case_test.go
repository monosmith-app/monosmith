//go:build e2e

package goldenpath

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestExternalLanguagePlugin adds a language smith knows nothing about, kiwi,
// through an external plugin alone (testdata/kiwiplugin): projects, graph
// edges, tasks and cache, a published stack, CI setup, doctor and release.
func TestExternalLanguagePlugin(t *testing.T) {
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	plugin := filepath.Join(base, "kiwiplugin")
	if runtime.GOOS == "windows" {
		plugin += ".exe"
	}
	wd, _ := os.Getwd()
	run(t, filepath.Dir(wd), "go", "build", "-o", plugin, "./goldenpath/testdata/kiwiplugin")

	w.smith(true, "new", "acme", "--ci", "github,azure")
	w.dir = filepath.Join(base, "acme")
	// Installed on the plugin search path (the workspace's .smith/plugins), and
	// declared by name and version only.
	w.write(".smith/plugins/kiwi/plugin.json", `{"name": "kiwi", "version": "1.0.0", "command": [`+quote(filepath.ToSlash(plugin))+`]}`)
	toml, _ := os.ReadFile(filepath.Join(w.dir, "smith.toml"))
	w.write("smith.toml", string(toml)+"\n[plugins.external.kiwi]\nversion = \"^2\"\n")
	w.trust("kiwi")
	if out := w.smith(false, "projects"); !strings.Contains(out, "wants version ^2, but the plugin is 1.0.0") {
		t.Fatalf("a version smith.toml does not accept must fail:\n%s", out)
	}
	w.write("smith.toml", string(toml)+"\n[plugins.external.kiwi]\nversion = \"^1\"\n")
	w.trust("kiwi")
	if out := w.smith(true, "plugin", "list"); !strings.Contains(out, "kiwi           1.0.0") {
		t.Fatalf("plugin list:\n%s", out)
	}
	if out := w.smith(true, "plugin", "check", "kiwi", "--generate"); !strings.Contains(out, "✔ the generated package records its stack") || strings.Contains(out, "✖") {
		t.Fatalf("plugin check:\n%s", out)
	}
	ignore, _ := os.ReadFile(filepath.Join(w.dir, ".gitignore"))
	w.write(".gitignore", string(ignore)+".kiwi-registry\n")

	// Projects of the new language, one with the stack the plugin publishes.
	w.smith(true, "add", "package", "fruit", "--lang", "kiwi")
	w.smith(true, "add", "app", "basket", "--stack", "kiwi-strict")
	manifest, _ := os.ReadFile(filepath.Join(w.dir, "apps", "basket", "kiwi.toml"))
	w.write("apps/basket/kiwi.toml", string(manifest)+"imports = \"kiwi:fruit\"\n")
	if !hasEdge(w.smith(true, "graph", "--format", "json"), "basket", "fruit") {
		t.Fatal("the plugin's imports must become a graph edge")
	}
	if out := w.smith(true, "stacks"); !strings.Contains(out, "kiwi-strict  (plugin kiwi)") || !strings.Contains(out, "projects   basket") {
		t.Fatalf("stacks:\n%s", out)
	}
	if out := w.smith(true, "explain", "basket:test"); !strings.Contains(out, "from plugin kiwi (stack kiwi-strict)") {
		t.Fatalf("the stack replaces test:\n%s", out)
	}

	// Tasks run in dependency order and come from the cache the second time.
	out := w.smith(true, "run", "build", "test")
	for _, want := range []string{"fruit:build", "basket:fmt", "kiwi tests passed --strict", "kiwi tests passed"} {
		if !strings.Contains(out, want) {
			t.Fatalf("run lacks %q:\n%s", want, out)
		}
	}
	if out := w.smith(true, "run", "build", "test"); strings.Contains(out, "▶") {
		t.Fatalf("second run executed tasks:\n%s", out)
	}

	// CI installs the language the way its plugin says; doctor asks it.
	w.smith(true, "ci", "generate")
	for _, f := range []string{".github/workflows/ci.yml", "azure-pipelines.yml"} {
		if body, _ := os.ReadFile(filepath.Join(w.dir, filepath.FromSlash(f))); !strings.Contains(string(body), "install-kiwi.sh") {
			t.Fatalf("%s does not install kiwi:\n%s", f, body)
		}
	}
	if out := w.smith(true, "doctor", "--verbose"); !strings.Contains(out, "kiwi project fruit has a kiwi.toml") || !strings.Contains(out, "external plugin kiwi is trusted") {
		t.Fatalf("doctor:\n%s", out)
	}

	// Release publishes through the plugin and tags the version.
	remote := filepath.Join(base, "remote.git")
	run(t, base, "git", "init", "-q", "--bare", remote)
	w.git("remote", "add", "origin", remote)
	w.git("add", "-A")
	w.git("commit", "-qm", "feat: kiwi projects")
	w.git("push", "-q", "origin", "main")
	w.smith(true, "release", "--yes")
	if registry, _ := os.ReadFile(filepath.Join(w.dir, ".kiwi-registry")); string(registry) != "fruit@0.1.0\n" {
		t.Fatalf("registry %q", registry)
	}
	if tags := w.git("ls-remote", "--tags", "origin"); !strings.Contains(tags, "fruit@0.1.0") {
		t.Fatalf("tags:\n%s", tags)
	}

	// A plugin of your own, in Go: scaffolded, conformant, then in use.
	w.smith(true, "add", "app", "hello-rules", "--stack", "go-smith-plugin")
	if out := w.smith(true, "run", "test", "-p", "hello-rules"); !strings.Contains(out, "hello-rules:test ✔") {
		t.Fatalf("the scaffolded plugin's conformance test:\n%s", out)
	}
	toml, _ = os.ReadFile(filepath.Join(w.dir, "smith.toml"))
	w.write("smith.toml", string(toml)+"\n[plugins.external.hello-rules]\ncommand = [\"go\", \"run\", \"./apps/hello-rules\"]\n\n[conventions.hello-rules]\nlanguages = [\"go\"]\n")
	w.trust("hello-rules")
	if out := w.smith(true, "run", "hello", "-p", "hello-rules"); !strings.Contains(out, "hello-rules:hello ✔") {
		t.Fatalf("the plugin's task:\n%s", out)
	}
}

func quote(s string) string { return `"` + s + `"` }
