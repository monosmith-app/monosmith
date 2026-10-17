package pruning

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// workspace is an npm workspace whose app uses two workspace packages, one
// of them through the other, and nothing from the registry, so npm needs no
// network. core builds into smith's dist/packages/core; util builds in its
// own directory, as an Nx-converted package does.
func workspace(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm is not installed")
	}
	root := t.TempDir()
	write(t, root, "package.json", `{ "name": "acme", "private": true, "workspaces": ["apps/*", "packages/*"] }`)
	write(t, root, "apps/api/package.json", `{ "name": "@acme/api", "version": "1.2.0", "type": "commonjs", "main": "../../dist/apps/api/main.js", "dependencies": { "@acme/core": "*" }, "devDependencies": { "@acme/util": "*" } }`)
	write(t, root, "packages/core/package.json", `{ "name": "@acme/core", "version": "0.1.0", "main": "../../dist/packages/core/index.js", "peerDependencies": { "@acme/util": "*" } }`)
	write(t, root, "packages/util/package.json", `{ "name": "@acme/util", "version": "0.2.0", "main": "dist/index.js" }`)
	run(t, root, "npm", "install", "--ignore-scripts", "--no-audit", "--no-fund")
	// The builds.
	write(t, root, "dist/apps/api/main.js", `console.log(require('@acme/core').greet())`)
	write(t, root, "dist/packages/core/index.js", `exports.greet = () => 'hello ' + require('@acme/util').name`)
	write(t, root, "packages/util/dist/index.js", `exports.name = 'util'`)
	return root
}

func run(t *testing.T, dir string, argv ...string) string {
	t.Helper()
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %v\n%s", strings.Join(argv, " "), err, out)
	}
	return string(out)
}

func TestPruneWritesAnOutputThatInstallsAndRunsOnItsOwn(t *testing.T) {
	root := workspace(t)
	res, err := Prune(Request{WorkspaceRoot: root, Root: "apps/api"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Out != "dist/apps/api" || strings.Join(res.Modules, ",") != "@acme/core,@acme/util" {
		t.Fatalf("result %+v", res)
	}
	out := filepath.Join(root, "dist", "apps", "api")
	var manifest struct {
		Main         string            `json:"main"`
		Dependencies map[string]string `json:"dependencies"`
	}
	raw, _ := os.ReadFile(filepath.Join(out, "package.json"))
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Main != "./main.js" || manifest.Dependencies["@acme/core"] != "file:./workspace_modules/@acme/core" {
		t.Errorf("package.json = %s", raw)
	}
	if _, err := os.Stat(filepath.Join(out, ".npmrc")); err == nil {
		t.Error(".npmrc left in the output")
	}
	// Away from the workspace, as in an image.
	deploy := t.TempDir()
	if err := copyTree(out, deploy); err != nil {
		t.Fatal(err)
	}
	run(t, deploy, "npm", "ci", "--omit=dev", "--ignore-scripts", "--no-audit", "--no-fund")
	if got := strings.TrimSpace(run(t, deploy, "node", ".")); got != "hello util" {
		t.Errorf("the app printed %q", got)
	}
}

func TestPruneWritesIntoAnotherBuildDirectory(t *testing.T) {
	root := workspace(t)
	write(t, root, "apps/api/dist/main.js", "")
	recorded := ""
	original := npm
	npm = func(dir string, args ...string) error { recorded = dir; return nil }
	t.Cleanup(func() { npm = original })
	res, err := Prune(Request{WorkspaceRoot: root, Root: "apps/api", Out: "apps/api/dist"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Out != "apps/api/dist" || recorded != filepath.Join(root, "apps", "api", "dist") {
		t.Fatalf("result %+v, npm ran in %s", res, recorded)
	}
	raw, _ := os.ReadFile(filepath.Join(root, "apps", "api", "dist", "package.json"))
	if !strings.Contains(string(raw), `"main": "../../dist/apps/api/main.js"`) {
		t.Errorf("a path outside smith's build directory is left as it is: %s", raw)
	}
	for _, f := range []string{"package-lock.json", "workspace_modules/@acme/core/package.json", "workspace_modules/@acme/util/dist/index.js"} {
		if _, err := os.Stat(filepath.Join(root, "apps", "api", "dist", filepath.FromSlash(f))); err != nil {
			t.Errorf("no %s", f)
		}
	}
}

func TestPruneNeedsTheBuild(t *testing.T) {
	root := t.TempDir()
	write(t, root, "package-lock.json", `{"packages":{}}`)
	if _, err := Prune(Request{WorkspaceRoot: root, Root: "apps/api"}); err == nil || !strings.Contains(err.Error(), "build apps/api first") {
		t.Errorf("err = %v", err)
	}
}
