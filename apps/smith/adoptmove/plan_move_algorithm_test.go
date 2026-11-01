package adoptmove

import (
	"reflect"
	"strings"
	"testing"
)

var nodeFiles = []string{".github/ci.yml", ".gitignore", "LICENSE", "README.md", "package-lock.json", "package.json", "src/index.js", "test/a.js", "tsconfig.json"}

func TestPlanMoveSplitsProjectFromRepositoryFiles(t *testing.T) {
	c := Candidate{Ecosystem: EcosystemNode, Manifest: "package.json", Name: "web"}
	p, err := PlanMove(c, "apps/web", nodeFiles)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"git mv package.json apps/web/package.json", "git mv src apps/web/src", "git mv test apps/web/test", "git mv tsconfig.json apps/web/tsconfig.json"}
	if !reflect.DeepEqual(p.Lines(), want) {
		t.Errorf("moves = %v, want %v", p.Lines(), want)
	}
	if !reflect.DeepEqual(p.Stays, []string{".github", ".gitignore", "LICENSE", "README.md", "package-lock.json"}) {
		t.Errorf("stays = %v", p.Stays)
	}
	if len(p.Notes) != 1 || !strings.Contains(p.Notes[0], "package-lock.json") {
		t.Errorf("notes = %v", p.Notes)
	}
}

func TestPlanMoveRefusesAnAmbiguousLayout(t *testing.T) {
	c := Candidate{Ecosystem: EcosystemNode, Manifest: "package.json", Name: "web"}
	cases := []struct {
		name  string
		files []string
		dir   string
		why   string
	}{
		{"manifest not tracked", []string{"README.md", "src/a.js"}, "apps/web", "not in the repository's files"},
		{"folder already holds files", append([]string{"apps/web/x.js"}, nodeFiles...), "apps/web", "already holds files"},
		{"folder is a file", append([]string{"apps"}, nodeFiles...), "apps/web", "is a file"},
		{"first part is a project folder", append([]string{"packages/x/y.js"}, nodeFiles...), "packages/web", "has a folder or file named packages"},
	}
	for _, tc := range cases {
		_, err := PlanMove(c, tc.dir, tc.files)
		if err == nil || !strings.Contains(err.Error(), tc.why) {
			t.Errorf("%s: err = %v, want it to contain %q", tc.name, err, tc.why)
		}
	}
}

func TestPlanMoveAllowsAFolderBesideRepositoryFiles(t *testing.T) {
	c := Candidate{Ecosystem: EcosystemGo, Manifest: "go.mod", Name: "x"}
	// docs stays at the root, so docs/x is a legal place even though docs exists.
	p, err := PlanMove(c, "docs/x", []string{"docs/a.md", "go.mod", "main.go"})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Moves) != 2 {
		t.Errorf("moves = %v", p.Moves)
	}
}
