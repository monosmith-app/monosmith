package doctorfix

import "testing"

func TestUnifiedDiff(t *testing.T) {
	before := "a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nl\n"
	after := "a\nb\nc\nD\ne\nf\ng\nh\ni\nj\nl\nk\n"
	want := "--- a/x\n+++ b/x\n" +
		"@@ -1,7 +1,7 @@\n a\n b\n c\n-d\n+D\n e\n f\n g\n" +
		"@@ -9,3 +9,4 @@\n i\n j\n l\n+k\n"
	if got := UnifiedDiff("x", before, after); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	if UnifiedDiff("x", "same\n", "same\n") != "" {
		t.Fatal("no change, no diff")
	}
	if got := UnifiedDiff(".gitignore", "", ".smith/\n"); got != "--- a/.gitignore\n+++ b/.gitignore\n@@ -1,0 +1,1 @@\n+.smith/\n" {
		t.Fatalf("a new file:\n%s", got)
	}
}
