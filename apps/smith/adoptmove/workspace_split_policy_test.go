package adoptmove

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestClassifyKeepsWorkspaceFilesAtTheRootAndSendsDocumentsToDoc(t *testing.T) {
	cases := []struct {
		eco, name string
		want      Role
	}{
		{EcosystemNode, "orders-api.code-workspace", RoleRepository},
		{EcosystemGo, "Mono.CODE-WORKSPACE", RoleRepository},
		{EcosystemNode, ".prettierrc", RoleRepository},
		{EcosystemNode, "eslint.config.mjs", RoleRepository},
		{EcosystemNode, ".eslintrc.json", RoleRepository},
		{EcosystemNode, "tsconfig.json", RoleProject},
		{EcosystemNode, "jest.config.js", RoleProject},
		{EcosystemGo, ".prettierrc", RoleProject},
		{EcosystemGo, ".golangci.yml", RoleRepository},
		{EcosystemPython, "ruff.toml", RoleRepository},
		{EcosystemRust, "rustfmt.toml", RoleRepository},
		{EcosystemNode, "MIGRATION_PLAN.md", RoleDocs},
		{EcosystemGo, "design.rst", RoleDocs},
		{EcosystemNode, "documentation", RoleDocs},
		{EcosystemNode, "README.md", RoleRepository},
		{EcosystemNode, "LICENSE", RoleRepository},
		{EcosystemNode, "doc", RoleRepository},
		{EcosystemPython, "requirements.txt", RoleProject},
	}
	for _, tc := range cases {
		if got := Classify(tc.eco, tc.name); got != tc.want {
			t.Errorf("Classify(%s, %s) = %v, want %v", tc.eco, tc.name, got, tc.want)
		}
	}
}

func TestPlanMoveSendsDocumentsToDoc(t *testing.T) {
	c := Candidate{Ecosystem: EcosystemNode, Manifest: "package.json", Name: "web"}
	files := append([]string{"MIGRATION_PLAN.md", "web.code-workspace", ".prettierrc"}, nodeFiles...)
	p, err := PlanMove(c, "apps/web", files)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.Docs, []Step{{From: "MIGRATION_PLAN.md", To: "doc/MIGRATION_PLAN.md"}}) {
		t.Errorf("docs = %v", p.Docs)
	}
	for _, stays := range []string{"web.code-workspace", ".prettierrc"} {
		found := false
		for _, s := range p.Stays {
			found = found || s == stays
		}
		if !found {
			t.Errorf("%s must stay at the root: %v", stays, p.Stays)
		}
	}
	if lines := p.Lines(); lines[len(lines)-1] != "git mv MIGRATION_PLAN.md doc/MIGRATION_PLAN.md" {
		t.Errorf("lines = %v", lines)
	}
	if _, err := PlanMove(c, "apps/web", append([]string{"doc/MIGRATION_PLAN.md"}, files...)); err == nil || !strings.Contains(err.Error(), "overwrite") {
		t.Errorf("a document that would overwrite one in doc must be refused: %v", err)
	}
}

func TestSortDependenciesAndSettle(t *testing.T) {
	root := t.TempDir()
	manifest := `{"devDependencies": {"typescript": "5", "eslint": "8", "supertest": "7", "nock": "13"}}`
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	var p Plan
	if err := SortDependencies(root, Candidate{Ecosystem: EcosystemNode}, &p); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.Hoist, []string{"typescript"}) || !reflect.DeepEqual(p.Replaced, []string{"eslint"}) || !reflect.DeepEqual(p.Unsure, []string{"nock", "supertest"}) {
		t.Fatalf("plan = %+v", p)
	}
	p.Settle([]string{"supertest"})
	if !reflect.DeepEqual(p.Hoist, []string{"supertest", "typescript"}) || !reflect.DeepEqual(p.Keep, []string{"nock"}) || p.Unsure != nil {
		t.Errorf("settled plan = %+v", p)
	}

	var goPlan Plan
	if err := SortDependencies(root, Candidate{Ecosystem: EcosystemGo}, &goPlan); err != nil || !reflect.DeepEqual(goPlan, Plan{}) {
		t.Errorf("a language with no dependency split leaves the plan alone: %+v %v", goPlan, err)
	}
}

func TestAskWorkspace(t *testing.T) {
	cases := []struct {
		input   string
		want    bool
		wantErr bool
	}{
		{"1\n", true, false},
		{"2\n", false, false},
		{"\n", false, false},
		{"x\n1\n", true, false},
		{"x\ny\nz\n", false, true},
		{"", false, true},
	}
	for _, tc := range cases {
		var out bytes.Buffer
		got, err := AskWorkspace("supertest", strings.NewReader(tc.input), &out)
		if got != tc.want || (err != nil) != tc.wantErr {
			t.Errorf("input %q: got %v, %v; want %v, error %v", tc.input, got, err, tc.want, tc.wantErr)
		}
		if !strings.Contains(out.String(), "supertest") {
			t.Errorf("input %q: the question does not name the package: %s", tc.input, out.String())
		}
	}
}
