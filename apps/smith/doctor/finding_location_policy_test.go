package doctor

import (
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/boundaries"
	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
	"github.com/monosmith-app/monosmith/apps/smith/derive"
	"github.com/monosmith-app/monosmith/apps/smith/doctorchecks"
	"github.com/monosmith-app/monosmith/apps/smith/overridelearn"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/recipedrift"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// assertLocated fails for each finding that names no project and no valid
// area, or that names both, so the TUI always knows where to file it.
func assertLocated(t *testing.T, what string, fs ...protocol.Finding) {
	t.Helper()
	if len(fs) == 0 {
		t.Errorf("%s: no finding to check", what)
	}
	for _, f := range fs {
		switch {
		case f.Project != "" && f.Area != "":
			t.Errorf("%s: %q has both project %q and area %q", what, f.Check, f.Project, f.Area)
		case f.Project == "" && !doctorchecks.ValidArea(f.Area):
			t.Errorf("%s: %q has no project and area %q is not valid", what, f.Check, f.Area)
		}
		if f.Project != "" && strings.HasPrefix(f.Problem, f.Project+" ") {
			t.Errorf("%s: %q repeats the project in its problem %q", what, f.Check, f.Problem)
		}
	}
}

func TestEveryHostMapperLocatesItsFindings(t *testing.T) {
	vs := []boundaries.Violation{{From: "a", To: "b", Rule: "r"}, {From: "a", To: "c", Rule: "r", Warning: true}}
	assertLocated(t, "boundaries", BoundaryFindings(vs)...)
	assertLocated(t, "boundaries ok", BoundaryFindings(nil)...)

	assertLocated(t, "groups", GroupFindings(nil, []workspace.Group{{}})...)

	cache := cachemaintenance.Status{Exists: true, Bytes: 10 << 20, LimitBytes: 1 << 20, BrokenEntries: 2}
	assertLocated(t, "local cache", localCacheFindings(cache)...)
	assertLocated(t, "cache key", cacheKeyFinding("s3://x", func(string) string { return "" })...)
	assertLocated(t, "defaults", defaultsVintageFinding("", ""), defaultsVintageFinding("2999.01", ""))

	now := time.Now()
	lib, _ := UnusedLibHint("core", 0, 2, now.Add(-30*24*time.Hour), now)
	pkgHint, _ := UnpublishedPackageHint("cli", true)
	assertLocated(t, "kind hints", lib, pkgHint)
	if lib.Project != "core" || pkgHint.Project != "cli" {
		t.Errorf("a kind hint is about its project: %q, %q", lib.Project, pkgHint.Project)
	}

	assertLocated(t, "recipe drift", doctorchecks.RecipeFinding(recipedrift.Report{Project: "orders", Recipe: "team", Problem: "no network"}))
	assertLocated(t, "recipe ok", doctorchecks.RecipeFinding(recipedrift.Report{Project: "orders", Recipe: "team", From: "1", To: "1"}))
	assertLocated(t, "overrides",
		overrideSuggestionFinding(overridelearn.Suggestion{Kind: overridelearn.Replace, Task: "test", Field: "cache", Value: "x", Projects: []string{"a"}, Total: 2}),
		overrideSuggestionFinding(overridelearn.Suggestion{Task: "test", Field: "cache", Value: "x", Projects: []string{"a", "b"}, Total: 4}))
	assertLocated(t, "derived", deriveFindings(derive.Plan{})...)

	pack := &protocol.Registries{PackToRelease: true}
	assertLocated(t, "github release", githubReleaseFindings([]*projectgraph.Project{pkg("core", "node", pack), pkg("lib", "python", nil)}, false, func(string) bool { return false }, func(string) string { return "" })...)
}

func TestACollapsedMachineCheckSpeaksForTheMachine(t *testing.T) {
	one := Finding{Finding: protocol.Finding{Check: "a JDK is installed", Project: "api", Detail: "is not installed"}}
	two := one
	two.Project = "web"
	got := withoutRepeatedMachineChecks([]Finding{one, two})
	if len(got) != 1 || got[0].Project != "" || got[0].Area != doctorchecks.AreaToolchain {
		t.Fatalf("shared by two projects it is about the machine: %+v", got)
	}
	got = withoutRepeatedMachineChecks([]Finding{one})
	if got[0].Project != "api" {
		t.Fatalf("a single project keeps its finding: %+v", got)
	}
}
