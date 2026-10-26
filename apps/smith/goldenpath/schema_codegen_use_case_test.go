//go:build e2e

package goldenpath

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const userProto = `syntax = "proto3";
package acme;
message User { string id = 1; }
`

const orderProto = `syntax = "proto3";
package acme;
message Order { string id = 1; int32 quantity = 2; }
`

// needProtoc is the real generator of these tests. CI installs it (the
// golden-path workflow); on a machine without it the test is skipped, but
// never in CI, where a missing generator must fail.
func needProtoc(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("protoc"); err != nil {
		if os.Getenv("CI") != "" {
			t.Fatal("protoc is not installed; the golden-path workflow must provide it")
		}
		t.Skip("protoc is not installed")
	}
}

// TestSchemaCodegen runs a real generator, protoc, through the generate tasks
// smith.toml declares: two schemas in one project, cached on the schema and
// restored from the cache, regenerated when a schema changes, and ordered
// before the tasks that read the code.
func TestSchemaCodegen(t *testing.T) {
	needProtoc(t)
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "acme", "--ci", "github")
	w.dir = filepath.Join(base, "acme")
	w.smith(true, "add", "lib", "pyclient", "--lang", "python")
	w.write("contracts/user.proto", userProto)
	w.write("contracts/order.proto", orderProto)
	version := strings.TrimSpace(strings.TrimPrefix(runOutput(t, "protoc", "--version"), "libprotoc"))
	w.write("smith.toml", w.read("smith.toml")+`
[projects.pyclient.generate]
schema = "contracts/user.proto"
generator = "python"
version = "`+version+`"
out = "src/pyclient/user_gen"

[[projects.pyclient.generates]]
schema = "contracts/order.proto"
generator = "python"
version = "`+version+`"
out = "src/pyclient/order_gen"
`)

	// The tasks are in the graph, ahead of the tasks that read their output.
	if out := w.smith(true, "explain", "pyclient:generate-order"); !strings.Contains(out, "protoc") || !strings.Contains(out, "order.proto") {
		t.Fatalf("explain:\n%s", out)
	}
	if out := w.smith(true, "explain", "pyclient:test"); !strings.Contains(out, "generate-order") || !strings.Contains(out, "generate") {
		t.Fatalf("test does not wait for the generate tasks:\n%s", out)
	}

	out := w.smith(true, "run", "generate", "generate-order", "-p", "pyclient")
	if !strings.Contains(out, "pyclient:generate ✔") || !strings.Contains(out, "pyclient:generate-order ✔") {
		t.Fatalf("run:\n%s", out)
	}
	user := w.read("libs/pyclient/src/pyclient/user_gen/user_pb2.py")
	order := w.read("libs/pyclient/src/pyclient/order_gen/order_pb2.py")
	if !strings.Contains(user, "User") || !strings.Contains(order, "Order") || strings.Contains(user, "Order") {
		t.Fatalf("generated code:\n%s\n%s", user, order)
	}

	// Nothing changed: both come from the cache, and a deleted output returns.
	if out := w.smith(true, "run", "generate", "generate-order", "-p", "pyclient"); strings.Contains(out, "▶") {
		t.Fatalf("second run executed a generator:\n%s", out)
	}
	if err := os.RemoveAll(filepath.Join(w.dir, "libs", "pyclient", "src", "pyclient", "order_gen")); err != nil {
		t.Fatal(err)
	}
	if out := w.smith(true, "run", "generate-order", "-p", "pyclient"); !strings.Contains(out, "from cache") {
		t.Fatalf("restore:\n%s", out)
	}
	if !strings.Contains(w.read("libs/pyclient/src/pyclient/order_gen/order_pb2.py"), "Order") {
		t.Fatal("the cache did not restore the generated code")
	}

	// Only the schema that changed runs again.
	w.write("contracts/order.proto", strings.Replace(orderProto, "quantity", "amount", 1))
	out = w.smith(true, "run", "generate", "generate-order", "-p", "pyclient")
	if !strings.Contains(out, "pyclient:generate-order ▶") || strings.Contains(out, "pyclient:generate ▶") {
		t.Fatalf("a changed schema must regenerate only its own task:\n%s", out)
	}
	if !strings.Contains(w.read("libs/pyclient/src/pyclient/order_gen/order_pb2.py"), "amount") {
		t.Fatal("the new schema's field is not in the generated code")
	}
	// The version is part of the cache key.
	w.write("smith.toml", strings.Replace(w.read("smith.toml"), `version = "`+version+`"
out = "src/pyclient/user_gen"`, `version = "`+version+`.0"
out = "src/pyclient/user_gen"`, 1))
	if out := w.smith(true, "run", "generate", "-p", "pyclient"); !strings.Contains(out, "pyclient:generate ▶") {
		t.Fatalf("a new generator version must regenerate:\n%s", out)
	}
}

// TestPluginContributedGenerate has an external plugin (testdata/kiwiplugin)
// contribute a generate task from its settings: smith orders it before the
// project's build, caches it, and links the project to the schema's owner.
func TestPluginContributedGenerate(t *testing.T) {
	needProtoc(t)
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	plugin := filepath.Join(base, "kiwiplugin")
	if runtime.GOOS == "windows" {
		plugin += ".exe"
	}
	wd, _ := os.Getwd()
	run(t, filepath.Dir(wd), "go", "build", "-o", plugin, "./goldenpath/testdata/kiwiplugin")

	w.smith(true, "new", "acme", "--ci", "github")
	w.dir = filepath.Join(base, "acme")
	w.write(".smith/plugins/kiwi/plugin.json", `{"name": "kiwi", "version": "1.0.0", "command": [`+quote(filepath.ToSlash(plugin))+`]}`)
	w.smith(true, "add", "app", "basket", "--stack", "kiwi-strict")
	w.smith(true, "add", "package", "catalog", "--lang", "kiwi")
	w.write("packages/catalog/schema/user.proto", userProto)
	w.write("smith.toml", w.read("smith.toml")+"\n[plugins.external.kiwi]\nversion = \"^1\"\n\n[projects.basket.settings.kiwi]\nschema = \"packages/catalog/schema/user.proto\"\n")

	if !hasEdge(w.smith(true, "graph", "--format", "json"), "basket", "catalog") {
		t.Fatal("the schema a plugin's generate task reads must link its project to the schema's owner")
	}
	if out := w.smith(true, "explain", "basket:build"); !strings.Contains(out, "generate-schema") {
		t.Fatalf("build does not wait for the plugin's generate task:\n%s", out)
	}
	out := w.smith(true, "run", "build", "-p", "basket")
	if !strings.Contains(out, "basket:generate-schema ✔") {
		t.Fatalf("run:\n%s", out)
	}
	if !strings.Contains(w.read("apps/basket/gen/user_pb2.py"), "User") {
		t.Fatal("the plugin's generate task wrote no code")
	}
	if out := w.smith(true, "run", "build", "-p", "basket"); strings.Contains(out, "▶") {
		t.Fatalf("second run executed tasks:\n%s", out)
	}
}

func runOutput(t *testing.T, name string, args ...string) string {
	t.Helper()
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
	return string(out)
}
