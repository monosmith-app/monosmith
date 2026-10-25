package cachecommand

import (
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
)

func TestStatusTextOfAFullCacheWithARemote(t *testing.T) {
	pruned := time.Date(2026, 10, 25, 9, 41, 0, 0, time.UTC)
	percent := 17.2
	got := statusText(StatusReport{
		Status: cachemaintenance.Status{
			Dir: "/work/acme/.smith/cache", Exists: true, Entries: 12, Blobs: 47, Bytes: 352 << 20, LimitBytes: 2048 << 20, LastPrune: &pruned,
			Largest: []cachemaintenance.TaskUsage{{Task: "web:build", Entries: 3, LargestBytes: 120 << 20}, {Task: "api:test", Entries: 1, LargestBytes: 2048}},
		},
		UsedPercent: &percent,
		Remote:      RemoteSummary{Configured: true, Scheme: "s3", Host: "acme-cache", Active: true, ReadOnly: true},
	})
	want := `local cache  /work/acme/.smith/cache
  entries     12
  blobs       47
  size        352.0 MB of 2.0 GB (17.2% full)
  last prune  2026-10-25 09:41
  remote      s3://acme-cache (active, read-only)
largest entries
  web:build                  120.0 MB  3 entries
  api:test                     2.0 KB  1 entry
`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestStatusTextOfACacheNobodyUsedYet(t *testing.T) {
	got := statusText(StatusReport{Status: cachemaintenance.Status{Dir: "/w/.smith/cache"}})
	want := `local cache  /w/.smith/cache
  entries     none yet: nothing has been cached
  size        0 B (no limit: [cache] max_size_mb is 0)
  last prune  never
  remote      none
`
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestStatusTextNamesTheRemedyForEveryProblem(t *testing.T) {
	percent := 150.0
	got := statusText(StatusReport{
		Status:      cachemaintenance.Status{Dir: "d", Exists: true, Entries: 2, Blobs: 1, Bytes: 3 << 20, LimitBytes: 2 << 20, BrokenEntries: 1},
		UsedPercent: &percent,
		Remote:      RemoteSummary{Configured: true, Scheme: "https", Host: "cache.example.com"},
	})
	for _, want := range []string{
		"✖ the cache is over its limit\n    → smith cache prune brings it under 2.0 MB",
		"✖ 1 entry cannot be used: unreadable, or naming a blob that is gone\n    → smith cache verify --fix removes them",
		"https://cache.example.com (off: set SMITH_CACHE_KEY to use it)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestPruneTextCases(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   cachemaintenance.PruneOutcome
		want string
	}{
		{"no cache", cachemaintenance.PruneOutcome{Dir: "/c"}, "The cache at /c is empty: nothing to prune.\n"},
		{"nothing to do", cachemaintenance.PruneOutcome{Exists: true, LimitBytes: 2048 << 20, SizeAfter: 5 << 20, Kept: 4},
			"Nothing to prune: the cache is 5.0 MB in 4 entries, within its limit of 2.0 GB.\n"},
		{"no limit and nothing unused", cachemaintenance.PruneOutcome{Exists: true, SizeAfter: 1024, Kept: 1},
			"Nothing to prune: the cache is 1.0 KB in 1 entry, within no limit.\n"},
		{"pruned", cachemaintenance.PruneOutcome{Exists: true, LimitBytes: 10 << 20, Entries: 3, Blobs: 5, Freed: 12 << 20, SizeBefore: 20 << 20, SizeAfter: 8 << 20, Kept: 2, Tasks: map[string]int{"web:build": 2, "api:test": 1}},
			"Pruned the cache: removed 3 entries and 5 blobs (12.0 MB of 20.0 MB); 8.0 MB left in 2 entries.\n  web:build                2 entries\n  api:test                 1 entry\n"},
		{"dry run", cachemaintenance.PruneOutcome{Exists: true, DryRun: true, Entries: 1, Blobs: 1, Freed: 2048, SizeBefore: 4096, SizeAfter: 2048, Kept: 1, Tasks: map[string]int{"web:build": 1}},
			"Would remove 1 entry and 1 blob (2.0 KB of 4.0 KB); 2.0 KB would be left in 1 entry.\n  web:build                1 entry\n"},
		{"skipped", cachemaintenance.PruneOutcome{Exists: true, Skipped: 2, SizeAfter: 10, Kept: 1},
			"Nothing to prune: the cache is 10 B in 1 entry, within no limit.\n! 2 files could not be removed (another process has them open)\n  → run `smith cache prune` again when no build is running\n"},
	} {
		if got := pruneText(tc.in); got != tc.want {
			t.Errorf("%s:\ngot:  %q\nwant: %q", tc.name, got, tc.want)
		}
	}
}

func TestVerifyTextListsProblemsAndMarksTheFixedOnes(t *testing.T) {
	digest := strings.Repeat("ab", 32)
	got := verifyText(cachemaintenance.VerifyReport{Dir: "/c", Entries: 3, Blobs: 4, Problems: []cachemaintenance.Problem{
		{Kind: cachemaintenance.CorruptBlob, Key: digest, Detail: "its content hashes to 00", Fixed: true},
		{Kind: cachemaintenance.LeftoverTmp, Key: "put-1", Detail: "a write that did not finish"},
	}})
	want := "Checked 3 entries and 4 blobs in /c: 2 problems.\n" +
		"  corrupt-blob     abababababab…  its content hashes to 00 (removed)\n" +
		"  leftover-tmp     put-1          a write that did not finish\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
	clean := verifyText(cachemaintenance.VerifyReport{Dir: "/c", Entries: 1, Blobs: 2})
	if clean != "Checked 1 entry and 2 blobs in /c: no problems.\n" {
		t.Errorf("clean: %q", clean)
	}
}

func TestFormatSize(t *testing.T) {
	for n, want := range map[int64]string{0: "0 B", 1023: "1023 B", 1024: "1.0 KB", 1536: "1.5 KB", 5 << 20: "5.0 MB", 2048 << 20: "2.0 GB", 3 << 40: "3.0 TB"} {
		if got := formatSize(n); got != want {
			t.Errorf("formatSize(%d) = %q, want %q", n, got, want)
		}
	}
}
