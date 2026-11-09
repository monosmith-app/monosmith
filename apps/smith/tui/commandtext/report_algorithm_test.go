package commandtext

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestReportTargetReadsTheOutputFile(t *testing.T) {
	for argv, want := range map[string]string{"report -o a.html": "a.html", "report --output b.html": "b.html", "report": "", "doctor -o x": ""} {
		got, ok := ReportTarget(strings.Fields(argv))
		if got != want || ok != (want != "") {
			t.Errorf("%q: %q %v", argv, got, ok)
		}
	}
}

func TestReportPathNamesTheFileAfterTheWorkspace(t *testing.T) {
	if got := filepath.Base(ReportPath("my ws/1")); got != "smith-report-my-ws-1.html" {
		t.Fatalf("got %q", got)
	}
	if got := filepath.Base(ReportPath("///")); got != "smith-report-workspace.html" {
		t.Fatalf("got %q", got)
	}
}
