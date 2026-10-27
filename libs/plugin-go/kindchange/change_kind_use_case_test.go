package kindchange

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func edited(r protocol.ChangeKindResult) map[string]string {
	out := map[string]string{}
	for _, f := range r.Files {
		out[f.Path] = f.Content
	}
	return out
}

func projects() []protocol.Project {
	return []protocol.Project{
		{Name: "core", Root: "libs/core", Language: "go"},
		{Name: "util", Root: "libs/util", Language: "go"},
		{Name: "api", Root: "apps/api", Language: "go"},
	}
}

func workspace() map[string]string {
	return map[string]string{
		"go.work": "go 1.26.0\n\nuse (\n\t./apps/api\n\t./libs/core\n\t./libs/util\n)\n",
		"libs/core/go.mod": "// smith:kind lib\nmodule example.com/ws/libs/core\n\ngo 1.26.0\n\n" +
			"require example.com/ws/libs/util v0.0.0\n\nreplace example.com/ws/libs/util => ../util\n",
		"libs/core/greeting/greet.go": "package greeting\n\nimport \"example.com/ws/libs/util/text\"\n\nvar _ = text.X\n",
		"libs/core/self_test.go":      "package core\n\nimport _ \"example.com/ws/libs/core/greeting\"\n",
		"libs/util/go.mod":            "// smith:kind lib\nmodule example.com/ws/libs/util\n\ngo 1.26.0\n",
		"libs/util/text/text.go":      "package text\n\nvar X = 1\n",
		"apps/api/go.mod": "// smith:kind app\nmodule example.com/ws/apps/api\n\ngo 1.26.0\n\n" +
			"require example.com/ws/libs/core v0.0.0\n\nreplace example.com/ws/libs/core => ../../libs/core\n",
		"apps/api/main.go": "package main\n\nimport \"example.com/ws/libs/core/greeting\"\n\nfunc main() { _ = greeting.X }\n",
	}
}

func TestLibBecomesPackageAndTheModulePathFollowsTheDirectory(t *testing.T) {
	dir := writeFiles(t, workspace())
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: projects()[0], To: protocol.ProjectPackage, NewRoot: "packages/core", Projects: projects(),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := edited(r)
	core := got["libs/core/go.mod"]
	for _, want := range []string{"// smith:kind package\n", "module example.com/ws/packages/core\n", "require example.com/ws/libs/util v0.0.0", "=> ../../libs/util"} {
		if !strings.Contains(core, want) {
			t.Errorf("core go.mod lacks %q:\n%s", want, core)
		}
	}
	if _, touched := got["libs/core/greeting/greet.go"]; touched {
		t.Errorf("a file that imports only another module is not edited")
	}
	if !strings.Contains(got["libs/core/self_test.go"], `"example.com/ws/packages/core/greeting"`) {
		t.Errorf("an import of its own slice follows:\n%s", got["libs/core/self_test.go"])
	}
	if !strings.Contains(got["apps/api/main.go"], `"example.com/ws/packages/core/greeting"`) {
		t.Errorf("dependents' imports follow:\n%s", got["apps/api/main.go"])
	}
	api := got["apps/api/go.mod"]
	if !strings.Contains(api, "require example.com/ws/packages/core v0.0.0") || !strings.Contains(api, "replace example.com/ws/packages/core => ../../packages/core") {
		t.Errorf("dependents' go.mod follow:\n%s", api)
	}
	if !strings.Contains(got["go.work"], "./packages/core") || strings.Contains(got["go.work"], "./libs/core") {
		t.Errorf("go.work follows:\n%s", got["go.work"])
	}
	if _, touched := got["libs/util/go.mod"]; touched {
		t.Error("a module that does not use core is not edited")
	}
	if len(r.Warnings) != 2 {
		t.Errorf("the tag directories and the module path are named: %v", r.Warnings)
	}
}

func TestAProgramCannotBecomeALibWithoutSaying(t *testing.T) {
	dir := writeFiles(t, workspace())
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: projects()[2], To: protocol.ProjectLib, NewRoot: "libs/api", Projects: projects(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(r.Warnings, "\n"), "package main") {
		t.Fatalf("the program is named: %v", r.Warnings)
	}
	if !strings.Contains(edited(r)["apps/api/go.mod"], "module example.com/ws/libs/api") {
		t.Fatalf("module path follows: %v", edited(r)["apps/api/go.mod"])
	}
}

func TestAModuleWhosePathDoesNotEndInItsDirectoryKeepsItsPath(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"libs/core/go.mod": "module example.com/core\n\ngo 1.26.0\n",
	})
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: protocol.Project{Name: "core", Root: "libs/core", Language: "go"},
		To:      protocol.ProjectPackage, NewRoot: "packages/core",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := edited(r)["libs/core/go.mod"]
	if got != "// smith:kind package\nmodule example.com/core\n\ngo 1.26.0\n" {
		t.Fatalf("only the kind changes:\n%s", got)
	}
}
