package fixwizard

import (
	"strings"
	"testing"
)

func TestSummarizeDiffsGivesOneLinePerFileAndKeepsTheRest(t *testing.T) {
	in := strings.Split(strings.TrimRight(`● Move the config
--- a/a.json
+++ b/a.json
@@ -1,2 +1,2 @@
-old
+new
  step 9: looks like a step
--- a/turbo.json
+++ b/turbo.json
@@ -1,2 +0,0 @@
-x
-y
--- a/b.json
+++ b/b.json
@@ -0,0 +1,1 @@
+z
  step 1: $ pnpm install`, "\n"), "\n")
	got, found := summarizeDiffs(in)
	want := []string{
		"● Move the config",
		"  a.json  changed  +1 -1",
		"  turbo.json  removed  +0 -2",
		"  b.json  new file  +1 -0",
		"  step 1: $ pnpm install",
	}
	if !found || strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("got %q (found %v), want %q", got, found, want)
	}
}

func TestSummarizeDiffsOfATextWithoutDiffsIsTheText(t *testing.T) {
	in := []string{"● x", "  step 1: $ y"}
	got, found := summarizeDiffs(in)
	if found || strings.Join(got, "|") != strings.Join(in, "|") {
		t.Fatalf("got %q, found %v", got, found)
	}
}
