package doctor

import "testing"

func TestTheRemoteCacheKeyIsOnlyAWarning(t *testing.T) {
	none := func(string) string { return "" }
	if got := cacheKeyFinding("", none); len(got) != 0 {
		t.Errorf("no remote, no finding: %+v", got)
	}
	got := cacheKeyFinding("azblob://acme/smith-cache/main", none)
	if len(got) != 1 || got[0].OK || !got[0].Warning {
		t.Fatalf("a remote without its key warns and does not fail: %+v", got)
	}
	if got := cacheKeyFinding("azblob://acme/smith-cache/main", func(string) string { return "$(SMITH_CACHE_KEY)" }); got[0].OK {
		t.Error("an unexpanded Azure macro is no key")
	}
	if got := cacheKeyFinding("azblob://acme/smith-cache/main", func(name string) string {
		if name == "SMITH_CACHE_KEY" {
			return "a-shared-secret"
		}
		return ""
	}); !got[0].OK {
		t.Errorf("with the key set the check passes: %+v", got)
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
