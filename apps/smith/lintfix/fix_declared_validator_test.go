package lintfix

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	bicep "github.com/monosmith-app/monosmith/libs/plugin-bicep/discovery"
	cpp "github.com/monosmith-app/monosmith/libs/plugin-cpp/discovery"
	dart "github.com/monosmith-app/monosmith/libs/plugin-dart/discovery"
	dotnet "github.com/monosmith-app/monosmith/libs/plugin-dotnet/discovery"
	golang "github.com/monosmith-app/monosmith/libs/plugin-go/discovery"
	golint "github.com/monosmith-app/monosmith/libs/plugin-go/linting"
	jvm "github.com/monosmith-app/monosmith/libs/plugin-jvm/discovery"
	lua "github.com/monosmith-app/monosmith/libs/plugin-lua/discovery"
	node "github.com/monosmith-app/monosmith/libs/plugin-node/discovery"
	objc "github.com/monosmith-app/monosmith/libs/plugin-objc/discovery"
	php "github.com/monosmith-app/monosmith/libs/plugin-php/discovery"
	python "github.com/monosmith-app/monosmith/libs/plugin-python/discovery"
	ruby "github.com/monosmith-app/monosmith/libs/plugin-ruby/discovery"
	rust "github.com/monosmith-app/monosmith/libs/plugin-rust/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// lintTool is the lint task a plugin describes, and the reason its tool has no
// fix mode, empty when the plugin declares one.
type lintTool struct {
	plugin  string
	task    protocol.Task
	noFixes string
}

// lintTools lists, for each plugin that describes a lint task, that task as
// the plugin builds it. A plugin whose tool has no fix mode says why.
func lintTools(t *testing.T) []lintTool {
	t.Helper()
	workspace := t.TempDir()
	goLint, err := golint.Command(workspace, workspace)
	if err != nil {
		t.Fatal(err)
	}
	app := protocol.ProjectApp
	return []lintTool{
		{"go", golang.Tasks("tool", "apps/tool", app, false, goLint)["lint"], ""},
		{"python", python.Tasks("apps/api", app, "api", python.ToolArgs{})["lint"], ""},
		{"rust", rust.Tasks("apps/cli", "cli", app, nil)["lint"], ""},
		{"dotnet", dotnet.Tasks("apps/api", "Api.csproj", "", app)["lint"], ""},
		{"dart", dart.Tasks(dart.Shape{Root: "apps/ui", Name: "ui", Kind: app})["lint"], ""},
		{"ruby", ruby.Tasks(app, "", false)["lint"], ""},
		{"node", node.TaskFor("lint", "apps/web", map[string]string{"lint": "eslint ."}), ""},
		{"jvm (ktlint)", jvm.GradleTasks("apps/api", jvm.GradleBuild{Lint: []string{"ktlintCheck"}}, app)["lint"], ""},
		{"jvm (Android Lint)", jvm.GradleTasks("apps/api", jvm.GradleBuild{Lint: []string{"lint"}, Android: true}, app)["lint"], ""},
		{"jvm (Checkstyle)", jvm.GradleTasks("apps/api", jvm.GradleBuild{Lint: []string{"checkstyleMain"}}, app)["lint"], "Checkstyle only reports; it has no fix mode"},
		{"jvm (Maven)", jvm.Tasks("apps/api", "g:a", "a", app)["lint"], "Checkstyle only reports; it has no fix mode"},
		{"bicep", bicep.Tasks("infra", false)["lint"], "az bicep lint only reports; it has no fix mode"},
		{"cpp", cpp.Tasks(cpp.Shape{Root: "apps/core", Kind: app, Lint: []string{"src"}})["lint"], "cppcheck only reports; it has no fix mode"},
		{"php", php.Tasks(php.Shape{Kind: app})["lint"], "PHPStan only reports; it has no fix mode"},
		{"lua", lua.Tasks("libs/x", "x.rockspec", false)["lint"], "luacheck only reports; it has no fix mode"},
		{"objc", objc.Tasks("X.podspec")["lint"], "pod lib lint only reports; it has no fix mode"},
	}
}

func TestEveryLintToolEitherDeclaresAFixOrNamesWhyItHasNone(t *testing.T) {
	for _, tool := range lintTools(t) {
		t.Run(tool.plugin, func(t *testing.T) {
			if len(tool.task.Command) == 0 {
				t.Fatalf("%s describes no lint task: remove it from the list or fix the sample", tool.plugin)
			}
			declared := tool.task.Fix != nil && len(tool.task.Fix.Command) > 0
			switch {
			case declared && tool.noFixes != "":
				t.Fatalf("%s declares a fix and is also listed with no fix mode (%s): remove the absence", tool.plugin, tool.noFixes)
			case !declared && tool.noFixes == "":
				t.Fatalf("%s has a lint tool with no fix: declare protocol.Task.Fix, or list the tool with the reason it has no fix mode", tool.plugin)
			}
			if declared && tool.task.Fix.Count != "" && tool.task.Fix.Count != protocol.FixCountESLintJSON {
				t.Fatalf("%s: unknown fix count %q", tool.plugin, tool.task.Fix.Count)
			}
		})
	}
}

// TestEveryPluginThatDescribesALintTaskIsInTheList fails when a plugin defines a
// "lint" task in its discovery code but is not in lintTools, so a new linter
// cannot skip the fix question.
func TestEveryPluginThatDescribesALintTaskIsInTheList(t *testing.T) {
	listed := map[string]bool{}
	for _, tool := range lintTools(t) {
		listed[strings.Fields(tool.plugin)[0]] = true
	}
	// Plugins that run smith's own checks as "lint" or do not describe a linter
	// themselves: their lint is another tool's, listed above.
	owners := map[string]string{
		"plugin-node-toolchains": "node",
		"plugin-vertical-slices": "node",
		"plugin-architectures":   "node",
		"plugin-recipe":          "",
		"plugin-bundle":          "",
		"plugin-applescript":     "",
		"plugin-vba":             "",
		"plugin-docker":          "",
		"plugin-swift":           "",
	}
	dirs, err := filepath.Glob(filepath.Join("..", "..", "..", "libs", "plugin-*"))
	if err != nil || len(dirs) == 0 {
		t.Fatalf("no plugin folders found: %v", err)
	}
	for _, dir := range dirs {
		name := filepath.Base(dir)
		if _, skip := owners[name]; skip {
			continue
		}
		if !definesLint(t, dir) {
			continue
		}
		if !listed[strings.TrimPrefix(name, "plugin-")] {
			t.Errorf("%s describes a lint task but is not in lintTools: add it, with its fix or the reason it has none", name)
		}
	}
}

// definesLint reports whether a discovery source file of the plugin builds a
// task named lint.
func definesLint(t *testing.T, dir string) bool {
	t.Helper()
	files, _ := filepath.Glob(filepath.Join(dir, "discovery", "*.go"))
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), `"lint":`) {
			return true
		}
	}
	return false
}
