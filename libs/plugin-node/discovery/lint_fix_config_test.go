package discovery

import (
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestALintScriptThatRunsESLintDeclaresAFixThatCountsItsReport(t *testing.T) {
	task := TaskFor("lint", "apps/web", map[string]string{"lint": "eslint . && prettier --check ."})
	if task.Fix == nil || task.Fix.Count != protocol.FixCountESLintJSON {
		t.Fatalf("fix = %+v", task.Fix)
	}
	if !slices.Contains(task.Fix.Command, "--fix") || slices.Contains(task.Fix.Probe, "--fix") || len(task.Fix.Probe) == 0 {
		t.Fatalf("the fix fixes and the probe only reports: %+v", task.Fix)
	}
}

func TestALintScriptWithoutESLintOrAnotherScriptDeclaresNoFix(t *testing.T) {
	if task := TaskFor("lint", "apps/web", map[string]string{"lint": "biome check ."}); task.Fix != nil {
		t.Fatalf("biome lint has no declared fix: %+v", task.Fix)
	}
	if task := TaskFor("test", "apps/web", map[string]string{"test": "eslint ."}); task.Fix != nil {
		t.Fatalf("only the lint script fixes: %+v", task.Fix)
	}
}
