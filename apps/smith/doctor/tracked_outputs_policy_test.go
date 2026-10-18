package doctor

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestTrackedGeneratedNamesIgnoredAndOutputFiles(t *testing.T) {
	web := &projectgraph.Project{Project: protocol.Project{Name: "web", Root: "apps/web", Tasks: map[string]protocol.Task{
		"build": {Outputs: []string{"../../dist/apps/web"}},
		"gen":   {Outputs: []string{"src/generated"}},
	}}}
	ignored := []string{"packages/eslint-config/tsconfig.typecheck.tsbuildinfo"}
	tracked := []string{"apps/web/src/main.ts", "apps/web/src/generated/api.ts", "README.md"}
	got := TrackedGenerated(ignored, tracked, []*projectgraph.Project{web})
	want := []string{
		"apps/web/src/generated/api.ts (inside web:gen's outputs)",
		"packages/eslint-config/tsconfig.typecheck.tsbuildinfo (tracked, but .gitignore ignores it)",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
	if got := TrackedGenerated(nil, []string{"apps/web/src/main.ts"}, []*projectgraph.Project{web}); len(got) != 0 {
		t.Errorf("a source file is not generated: %v", got)
	}
}
