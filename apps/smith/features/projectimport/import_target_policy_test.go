package projectimport

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/adoptmove"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestIsGitURLTellsURLsFromFolders(t *testing.T) {
	for arg, want := range map[string]bool{
		"https://github.com/owner/repo":     true,
		"git@github.com:owner/repo.git":     true,
		"ssh://git@host/owner/repo":         true,
		"file:///tmp/repo":                  true,
		"../my-app":                         false,
		`C:\Users\me\code\app`:              false,
		"apps/web":                          false,
		"https-tools/that-is-a-folder-name": false,
	} {
		if got := IsGitURL(arg); got != want {
			t.Errorf("IsGitURL(%q) = %v, want %v", arg, got, want)
		}
	}
}

func TestDefaultKindFollowsTheIdentification(t *testing.T) {
	if DefaultKind(adoptmove.Candidate{App: true, Published: true}) != protocol.ProjectApp {
		t.Error("a program is an app")
	}
	if DefaultKind(adoptmove.Candidate{Published: true}) != protocol.ProjectPackage {
		t.Error("a published project is a package")
	}
	if DefaultKind(adoptmove.Candidate{}) != protocol.ProjectLib {
		t.Error("the rest are libs")
	}
}

func TestFolderSettlesTheFolderAndTheKind(t *testing.T) {
	cases := []struct {
		kind       protocol.ProjectKind
		into       string
		dir        string
		wantKind   protocol.ProjectKind
		wantRefuse bool
	}{
		{kind: protocol.ProjectApp, dir: "apps/api", wantKind: protocol.ProjectApp},
		{kind: protocol.ProjectPackage, dir: "packages/api", wantKind: protocol.ProjectPackage},
		{into: "libs/shared-api", dir: "libs/shared-api", wantKind: protocol.ProjectLib},
		{kind: protocol.ProjectApp, into: "tools/api", dir: "tools/api", wantKind: protocol.ProjectApp},
		{into: "tools/api", wantRefuse: true},
		{kind: protocol.ProjectApp, into: "packages/api", wantRefuse: true},
		{into: "../outside", wantRefuse: true},
	}
	for _, c := range cases {
		dir, kind, err := Folder(c.kind, "api", c.into)
		if c.wantRefuse {
			if err == nil {
				t.Errorf("Folder(%q, %q) = %q, want a refusal", c.kind, c.into, dir)
			}
			continue
		}
		if err != nil || dir != c.dir || kind != c.wantKind {
			t.Errorf("Folder(%q, %q) = %q %q %v, want %q %q", c.kind, c.into, dir, kind, err, c.dir, c.wantKind)
		}
	}
}

func TestHistoryDefaultsByTheSourceAndRefusesWhatItCannotGive(t *testing.T) {
	if got, _ := History("", true); got != HistoryKeep {
		t.Errorf("a git source keeps its history by default, got %q", got)
	}
	if got, _ := History("", false); got != HistoryCopy {
		t.Errorf("a plain folder is copied, got %q", got)
	}
	if got, err := History(HistoryCopy, true); got != HistoryCopy || err != nil {
		t.Errorf("a git source may be copied: %q %v", got, err)
	}
	for _, flag := range []string{HistoryKeep, HistorySquash} {
		if _, err := History(flag, false); err == nil {
			t.Errorf("--history %s on a plain folder must be refused", flag)
		}
	}
	if _, err := History("all", true); err == nil {
		t.Error("an unknown mode must be refused")
	}
}

func TestLockfilesOfEachLanguage(t *testing.T) {
	if got := Lockfiles("python"); !reflect.DeepEqual(got, []string{"uv.lock"}) {
		t.Errorf("python: %v", got)
	}
	if got := Lockfiles("node"); len(got) == 0 {
		t.Error("node drops npm's, yarn's and pnpm's lockfiles")
	}
	if got := Lockfiles("dotnet"); len(got) != 0 {
		t.Errorf("dotnet has no workspace lockfile: %v", got)
	}
}
