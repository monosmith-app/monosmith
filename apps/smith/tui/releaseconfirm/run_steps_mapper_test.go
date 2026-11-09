package releaseconfirm

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/release"
)

func TestEachPackageIsWaitingBuildingOrTaggedAsItsOutputSays(t *testing.T) {
	items := []release.Item{{Project: "web", Tag: "web@1.1.0"}, {Project: "kit", Tag: "kit@0.1.0"}, {Project: "sdk", Tag: "sdk@2.0.0"}}
	out := []string{"web: building 1.1.0", "web:build ✓ 145ms", "web: tagged web@1.1.0", "kit: building 0.1.0"}
	got := strings.Join(Steps(plain, items, out, 0, false, false), "\n")
	for _, want := range []string{"✓ web  tagged web@1.1.0", "kit  building", "○ sdk  waiting"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q:\n%s", want, got)
		}
	}
}

func TestAPackageThatWasBuildingWhenTheReleaseFailedIsStopped(t *testing.T) {
	items := []release.Item{{Project: "kit", Tag: "kit@0.1.0"}}
	got := strings.Join(Steps(plain, items, []string{"kit: building 0.1.0"}, 0, true, false), "\n")
	if !strings.Contains(got, "× kit  stopped") {
		t.Fatalf("the failed step is not marked:\n%s", got)
	}
}

func TestAnOutputLineOfAnotherPackageWithTheSamePrefixIsNotCounted(t *testing.T) {
	items := []release.Item{{Project: "web", Tag: "web@1"}}
	got := strings.Join(Steps(plain, items, []string{"web-admin: tagged web-admin@1"}, 0, false, false), "\n")
	if !strings.Contains(got, "waiting") {
		t.Fatalf("web took the step of web-admin:\n%s", got)
	}
}
