package verifyfacts

import (
	"os"
	"path/filepath"
	"testing"
)

func put(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFindCoverageReadsTheReportsAProjectAlreadyHas(t *testing.T) {
	ws := t.TempDir()
	put(t, filepath.Join(ws, "apps", "web", "coverage", "coverage-summary.json"), `{"total":{"lines":{"pct":81.25}}}`)
	put(t, filepath.Join(ws, "libs", "py", "coverage.xml"), `<coverage line-rate="0.5" branch-rate="0"></coverage>`)
	put(t, filepath.Join(ws, "libs", "rs", "lcov.info"), "SF:a.rs\nLF:10\nLH:8\nend_of_record\nSF:b.rs\nLF:10\nLH:2\nend_of_record\n")
	put(t, filepath.Join(ws, "libs", "go", "coverage.out"), "mode: set\nexample.com/a.go:1.1,2.2 4 1\nexample.com/a.go:3.1,4.2 4 0\n")
	put(t, filepath.Join(ws, "libs", "net", "TestResults", "abc", "coverage.cobertura.xml"), `<coverage line-rate="0.9"></coverage>`)
	cases := map[string]struct {
		pct  float64
		path string
	}{
		"apps/web": {81.25, "apps/web/coverage/coverage-summary.json"},
		"libs/py":  {50, "libs/py/coverage.xml"},
		"libs/rs":  {50, "libs/rs/lcov.info"},
		"libs/go":  {50, "libs/go/coverage.out"},
		"libs/net": {90, "libs/net/TestResults/abc/coverage.cobertura.xml"},
	}
	for project, want := range cases {
		got, ok := FindCoverage(ws, project)
		if !ok || got.Path != want.path || got.Percent < want.pct-0.001 || got.Percent > want.pct+0.001 {
			t.Errorf("%s: got %+v, %v; want %v at %s", project, got, ok, want.pct, want.path)
		}
	}
	if _, ok := FindCoverage(ws, "apps/none"); ok {
		t.Error("a project with no report has coverage: smith never turns it on")
	}
}
