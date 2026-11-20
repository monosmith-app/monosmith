package projectimport

import (
	"strings"
	"testing"
)

func TestSummaryShowsTheFolderTheHistoryAndTheCommits(t *testing.T) {
	p := Plan{Name: "api", Kind: "app", Folder: "apps/api", Ecosystem: "node", Manifest: "package.json", History: HistoryKeep,
		Source: Source{Arg: "https://example.com/api", Git: true, Ref: "main", Commit: "abc1234", Commits: 9}, Lockfiles: []string{"package-lock.json"}}
	text := Summary(p)
	for _, want := range []string{"Import api from https://example.com/api", "a node app (package.json)", "apps/api", "the 9 commit(s) of main (abc1234)", "package-lock.json is dropped", "1 merge commit", "--commit"} {
		if !strings.Contains(text, want) {
			t.Errorf("the summary lacks %q:\n%s", want, text)
		}
	}
	if plan := Changes(p); len(plan.Groups) != 3 {
		t.Errorf("the plan has a group for the place, the lockfile and the manifest: %+v", plan.Groups)
	}
}

func TestSummaryOfAPlainFolderSaysNothingIsCommitted(t *testing.T) {
	text := Summary(Plan{Name: "lib", Kind: "lib", Folder: "libs/lib", Ecosystem: "python", History: HistoryCopy, Source: Source{Arg: "../lib"}})
	if !strings.Contains(text, "a plain folder is copied") || !strings.Contains(text, "smith commits nothing") {
		t.Errorf("%s", text)
	}
}

func TestOutcomeIsOneLeftToDoList(t *testing.T) {
	p := Plan{Name: "api", Folder: "apps/api"}
	if text := Outcome(p, nil); !strings.Contains(text, "imported api into apps/api, and smith doctor finds nothing to fix") {
		t.Errorf("%s", text)
	}
	text := Outcome(p, []Left{{Problem: "tests fail", Remedy: "fix them\nor skip them"}})
	for _, want := range []string{"imported api into apps/api. Left to do (1):", "  × tests fail", "    → fix them", "      or skip them"} {
		if !strings.Contains(text, want) {
			t.Errorf("the outcome lacks %q:\n%s", want, text)
		}
	}
}

func TestSummaryOfASourceWithoutManifestSaysWhatIsWritten(t *testing.T) {
	p := Plan{Name: "sorters", Kind: "lib", Folder: "libs/sorters", Ecosystem: "dotnet", Manifest: "a .csproj file", WriteManifest: "dotnet", Sources: 22, History: HistoryCopy, Source: Source{Arg: "../sorters"}}
	text := Summary(p)
	for _, want := range []string{"22 dotnet source file(s) and no manifest", "manifest write a .csproj file", "the sources stay where they are"} {
		if !strings.Contains(text, want) {
			t.Errorf("the summary lacks %q:\n%s", want, text)
		}
	}
	plan := Changes(p)
	if len(plan.Groups) < 2 || plan.Groups[1].Lines[0].Path != "libs/sorters/*.csproj" {
		t.Errorf("the plan shows the manifest it writes: %+v", plan.Groups)
	}
}

func TestSummaryOfASubFolderNamesItAndItsRef(t *testing.T) {
	p := Plan{Name: "scrutor", Kind: "package", Folder: "packages/scrutor", Ecosystem: "dotnet", Manifest: "Scrutor.csproj", History: HistoryKeep,
		Source: Source{Arg: "../scrutor", Git: true, Path: "src/Scrutor", PathRef: "master", Ref: pathBranch, Commit: "c87fc2f", Commits: 264}}
	text := Summary(p)
	for _, want := range []string{"a dotnet package (src/Scrutor/Scrutor.csproj)", "the 264 commit(s) of src/Scrutor at master (c87fc2f)"} {
		if !strings.Contains(text, want) {
			t.Errorf("the summary lacks %q:\n%s", want, text)
		}
	}
}
