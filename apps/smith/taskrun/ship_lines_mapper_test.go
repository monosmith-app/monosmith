package taskrun

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/shiplink"
)

func TestARunningProgramIsAWarningThatNamesTheProgramToClose(t *testing.T) {
	got := strings.Join(shipLines("api:build", shiplink.Report{Folder: "dist/apps/api", InUse: []string{"dist/apps/api/api.exe"}}), "\n")
	want := "api:build ▲ api.exe is in use, so dist/apps/api keeps its old copy. The new build is in the tool's folder.\n" +
		"api:build → close api.exe and build again"
	if got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}

func TestShipLinesCountTheFilesAndSayWhenNothingMatched(t *testing.T) {
	got := shipLines("api:build", shiplink.Report{Folder: "dist/apps/api", Shipped: 2, Copied: 1})
	if len(got) != 2 || got[0] != "api:build │ shipped 2 files to dist/apps/api" || !strings.HasPrefix(got[1], "api:build │ 1 of them is a copy") {
		t.Fatalf("lines %q", got)
	}
	none := shipLines("api:build", shiplink.Report{Folder: "dist/apps/api", NothingMatched: true})
	if len(none) != 2 || !strings.Contains(none[0], "made no file to ship") || !strings.HasPrefix(none[1], "api:build → run smith doctor") {
		t.Fatalf("lines %q", none)
	}
}
