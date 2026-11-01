package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestSourceConditionFindingsExplainTheTsNodeWorkaround(t *testing.T) {
	cfg := workspace.Config{TaskGroups: []workspace.TaskGroup{{
		Select: []string{"lang:node"},
		Tasks: map[string]workspace.TaskOverride{"test": {Env: map[string]string{
			tsNodeOptionsEnv: `{"moduleResolution":"node10", "module":"commonjs", "customConditions": null}`,
		}}},
	}}}
	got := SourceConditionFindings(cfg)
	if len(got) != 1 || !got[0].OK || !strings.Contains(got[0].Detail, "customConditions null") || !strings.Contains(got[0].Detail, "lang:node") {
		t.Fatalf("findings %+v", got)
	}
	cfg.TaskGroups[0].Tasks["test"] = workspace.TaskOverride{Env: map[string]string{tsNodeOptionsEnv: `{"module":"commonjs"}`}}
	if got := SourceConditionFindings(cfg); len(got) != 0 {
		t.Fatalf("options without the reset give no finding: %+v", got)
	}
}
