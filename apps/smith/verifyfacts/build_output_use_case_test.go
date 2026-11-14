package verifyfacts

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFindResultsAndFindBuildMeasureWhatIsOnDisk(t *testing.T) {
	ws := t.TempDir()
	put(t, filepath.Join(ws, "apps", "web", "test-results", "junit.xml"), "<testsuites/>")
	if got, ok := FindResults(ws, "apps/web"); !ok || got != "apps/web/test-results/junit.xml" {
		t.Errorf("FindResults = %q, %v", got, ok)
	}
	if _, ok := FindResults(ws, "apps/other"); ok {
		t.Error("results where there are none")
	}
	put(t, filepath.Join(ws, "dist", "apps", "web", "index.js"), strings.Repeat("x", 3000))
	put(t, filepath.Join(ws, "dist", "apps", "web", "assets", "a.css"), strings.Repeat("y", 1000))
	got, ok := FindBuild(ws, "dist/apps/web")
	if !ok || got.Path != "dist/apps/web" || got.Bytes != 4000 {
		t.Errorf("FindBuild = %+v, %v", got, ok)
	}
	if _, ok := FindBuild(ws, "dist/apps/none"); ok {
		t.Error("a build that was not written has a size")
	}
	for n, want := range map[int64]string{980: "980 B", 4000: "3.9 KB", 4 << 20: "4.0 MB", 3 << 30: "3.0 GB"} {
		if got := FormatSize(n); got != want {
			t.Errorf("FormatSize(%d) = %q, want %q", n, got, want)
		}
	}
}
