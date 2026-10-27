package taskfolding

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

const sample = `schema = 1

# The workspace.
[workspace]
name = "acme"

[tasks.lint]
cache = false

# Every library builds into dist.
[projects.a.tasks.build]
outputs = ["dist"]

[projects.a.tasks.test]
depends_on = ["^build"]
env = { MODE = "ci" }

[projects."b.lib".tasks.build]
outputs = ["dist"]

[projects."b.lib".tasks.test]
depends_on = ["^build"]
env = { MODE = "ci" }

[projects.odd]
tag_prefix = "odd-v" # keep my tags

[projects.odd.tasks.build]
outputs = ["dist"]

[projects.odd.tasks.test]
env = { MODE = "local" }
`

func sampleProjects(t *testing.T) []Project {
	t.Helper()
	cfg, err := workspace.Parse([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	var projects []Project
	for _, name := range []string{"a", "b.lib", "odd"} {
		p := lib(name, cfg.Projects[name].Tasks)
		projects = append(projects, p)
	}
	return projects
}

func TestRewriteKeepsCommentsAndWritesTheGroups(t *testing.T) {
	projects := sampleProjects(t)
	res := Fold(projects)
	got, ok := Rewrite(sample, projects, res)
	if !ok {
		t.Fatalf("not rewritten:\n%s", got)
	}
	for _, keep := range []string{"# The workspace.", "[tasks.lint]", `tag_prefix = "odd-v" # keep my tags`, `[projects.odd.tasks.test]`, `env = { MODE = "local" }`} {
		if !strings.Contains(got, keep) {
			t.Errorf("lost %q:\n%s", keep, got)
		}
	}
	for _, gone := range []string{"[projects.a.tasks", `[projects."b.lib".tasks`, "# Every library builds into dist.", "[projects.odd.tasks.build]"} {
		if strings.Contains(got, gone) {
			t.Errorf("still has %q:\n%s", gone, got)
		}
	}
	for _, added := range []string{"[[task_groups]]", `select = ["lang:node", "kind:lib"]`, "[task_groups.tasks.build]", "[task_groups.tasks.test]"} {
		if !strings.Contains(got, added) {
			t.Errorf("missing %q:\n%s", added, got)
		}
	}
	// odd's test keeps its env and gains the group's depends_on.
	cfg, err := workspace.Parse([]byte(got))
	if err != nil {
		t.Fatal(err)
	}
	if o := cfg.Projects["odd"].Tasks["test"]; o.Env["MODE"] != "local" || strings.Join(o.DependsOn, ",") != "^build" {
		t.Errorf("odd test = %+v", o)
	}
}

func TestRewriteGivesUpOnATaskSetInAnotherForm(t *testing.T) {
	src := strings.Replace(sample, "[projects.a.tasks.build]\noutputs = [\"dist\"]\n", "[projects.a]\ntasks.build = { outputs = [\"dist\"] }\n", 1)
	cfg, err := workspace.Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	var projects []Project
	for _, name := range []string{"a", "b.lib", "odd"} {
		projects = append(projects, lib(name, cfg.Projects[name].Tasks))
	}
	if _, ok := Rewrite(src, projects, Fold(projects)); ok {
		t.Error("a task outside its own table cannot be edited safely")
	}
}
