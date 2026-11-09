package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/overridelearn"
)

func TestOverrideSuggestionFindingsNameTheSettingAndTheFix(t *testing.T) {
	lift := overrideSuggestionFinding(overridelearn.Suggestion{Kind: overridelearn.Lift, Task: "test", Field: "cache", Value: "false", Projects: []string{"a", "b", "c"}, Total: 9})
	if !lift.Warning || !strings.Contains(lift.Detail, "3 projects set [tasks.test] cache = false (a, b, c)") || !strings.Contains(lift.Remedy, "[tasks.test]") {
		t.Errorf("lift: %+v", lift)
	}
	replace := overrideSuggestionFinding(overridelearn.Suggestion{Kind: overridelearn.Replace, Task: "build", Field: "cache", Value: "true", Projects: []string{"a", "b", "c", "d", "e", "f", "g"}, Total: 8})
	if !strings.Contains(replace.Detail, "7 of 8 projects override") || !strings.Contains(replace.Detail, "and 2 more") || !strings.Contains(replace.Remedy, "probably wrong") {
		t.Errorf("replace: %+v", replace)
	}
}
