package discovery

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-go/linting"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestBuildTagsJoinPlainAndPlatformLists(t *testing.T) {
	settings := map[string]any{
		"tags":         []any{"production", "gtk3"},
		"tags_linux":   []any{"gtk3", "webkit"},
		"tags_windows": []any{"win"},
	}
	for goos, want := range map[string][]string{
		"linux":   {"production", "gtk3", "webkit"},
		"windows": {"production", "gtk3", "win"},
		"darwin":  {"production", "gtk3"},
	} {
		got, err := buildTagsFor(settings, goos)
		if err != nil || !slices.Equal(got, want) {
			t.Errorf("%s: got %v, %v; want %v", goos, got, err, want)
		}
	}
}

func TestBuildTagsRefuseBadValues(t *testing.T) {
	for name, settings := range map[string]map[string]any{
		"a string":     {"tags": "production"},
		"a number":     {"tags": []any{1}},
		"a comma":      {"tags": []any{"a,b"}},
		"a space":      {"tags_linux": []any{"a b"}},
		"empty string": {"tags": []any{""}},
	} {
		if _, err := buildTagsFor(settings, "linux"); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

func TestTagsGoToEveryCompilingTask(t *testing.T) {
	tasks := Tasks("api", "apps/api", protocol.ProjectApp, false, linting.Lint{Command: []string{"golangci-lint", "run"}})
	withBuildTags(tasks, []string{"production", "gtk3"})
	for _, name := range []string{"build", "test", "package", "start"} {
		if got := tasks[name].Command[2]; got != "-tags=production,gtk3" {
			t.Errorf("%s: %v", name, tasks[name].Command)
		}
	}
	if got := strings.Join(tasks["lint"].Command, " "); got != "golangci-lint run --build-tags production,gtk3" {
		t.Errorf("lint: %s", got)
	}
}

func TestNoTagsLeavesTheTasksAlone(t *testing.T) {
	tasks := Tasks("api", "apps/api", protocol.ProjectApp, false, linting.Lint{})
	before := strings.Join(tasks["build"].Command, " ")
	withBuildTags(tasks, nil)
	if got := strings.Join(tasks["build"].Command, " "); got != before {
		t.Errorf("%s", got)
	}
}

func TestSettingsTagsReachAProjectsTasksAndABadValueNamesTheTable(t *testing.T) {
	ws := rootModule(t)
	res, err := DiscoverWithDeclared(ws, []string{"go.mod"}, []protocol.DeclaredProject{{Name: "mvd", Root: "apps/mvd"}},
		map[string]map[string]any{"mvd": {"tags": []any{"production"}}})
	if err != nil {
		t.Fatal(err)
	}
	if got := res.Projects[0].Tasks["package"].Command[2]; got != "-tags=production" {
		t.Errorf("package: %v", res.Projects[0].Tasks["package"].Command)
	}
	_, err = DiscoverWithDeclared(ws, []string{"go.mod"}, []protocol.DeclaredProject{{Name: "mvd", Root: "apps/mvd"}},
		map[string]map[string]any{"mvd": {"tags": "production"}})
	if err == nil || !strings.Contains(err.Error(), "[projects.mvd.settings.go]") {
		t.Errorf("error: %v", err)
	}
}
