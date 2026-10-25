package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
)

func TestNoCacheYetMeansNothingToCheck(t *testing.T) {
	if got := localCacheFindings(cachemaintenance.Status{}); got != nil {
		t.Fatalf("%+v", got)
	}
}

func TestAHealthyCacheWithinItsLimitPasses(t *testing.T) {
	got := localCacheFindings(cachemaintenance.Status{Exists: true, Bytes: 10 << 20, LimitBytes: 2048 << 20})
	if len(got) != 2 {
		t.Fatalf("%+v", got)
	}
	for _, f := range got {
		if !f.OK {
			t.Errorf("should pass: %+v", f)
		}
	}
	if got := localCacheFindings(cachemaintenance.Status{Exists: true, Bytes: 10 << 30}); !got[0].OK {
		t.Fatal("a cache with no limit is never over it")
	}
}

func TestACacheOverItsLimitWarnsAndNamesPrune(t *testing.T) {
	got := localCacheFindings(cachemaintenance.Status{Dir: "/w/.smith/cache", Exists: true, Bytes: 3 << 30, LimitBytes: 2 << 30})
	over := got[0]
	if over.OK || !over.Warning || !strings.Contains(over.Remedy, "smith cache prune") || !strings.Contains(over.Detail, "3.0 GB") || !strings.Contains(over.Detail, "2.0 GB") {
		t.Fatalf("%+v", over)
	}
}

func TestUnusableEntriesWarnAndNameVerifyFix(t *testing.T) {
	got := localCacheFindings(cachemaintenance.Status{Exists: true, BrokenEntries: 4})
	broken := got[1]
	if broken.OK || !broken.Warning || !strings.Contains(broken.Remedy, "smith cache verify --fix") || !strings.Contains(broken.Detail, "4 cache entries") {
		t.Fatalf("%+v", broken)
	}
}
