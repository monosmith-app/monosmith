package doctor

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
)

func TestLeadProblemNamesTheFirstAndCountsTheRest(t *testing.T) {
	cases := map[string][]string{"": nil, "a": {"a"}, "a (and 2 more)": {"a", "b", "c"}}
	for want, items := range cases {
		if got := leadProblem(items); got != want {
			t.Errorf("leadProblem(%v) = %q, want %q", items, got, want)
		}
	}
}

func TestProblemWhenIsEmptyForAPassingCheck(t *testing.T) {
	if problemWhen(false, "x") != "" || problemWhen(true, "x") != "x" {
		t.Error("problemWhen must keep the text only for a failure")
	}
}

func TestCacheKeyFindingNamesTheProblemOnlyWhenItFails(t *testing.T) {
	failed := cacheKeyFinding("s3://b", func(string) string { return "" })[0]
	if failed.OK || failed.Problem != "SMITH_CACHE_KEY is not set, so the remote cache s3://b is off" {
		t.Errorf("failed: %+v", failed)
	}
	passed := cacheKeyFinding("s3://b", func(string) string { return "k" })[0]
	if !passed.OK || passed.Problem != "" {
		t.Errorf("passed: %+v", passed)
	}
}

func TestLocalCacheFindingsNameWhatTheyFound(t *testing.T) {
	got := localCacheFindings(cachemaintenance.Status{Exists: true, BrokenEntries: 3})
	if got[0].Problem != "" || got[1].Problem != "3 local cache entries are unusable" {
		t.Errorf("%+v", got)
	}
}

func TestDefaultsVintageFindingNamesWhatItFound(t *testing.T) {
	f := defaultsVintageFinding("", "")
	if f.OK || f.Problem != "the workspace does not record which defaults it uses" {
		t.Errorf("%+v", f)
	}
}
