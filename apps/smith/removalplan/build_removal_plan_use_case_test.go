package removalplan

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
)

func workspace(files map[string]string) *legacyinventory.Tree {
	t := legacyinventory.NewTree("", nil)
	for name, content := range files {
		t.Write(name, content)
	}
	return t
}

var fullChoices = convertchoices.Choices{Level: convertchoices.LevelFull, Lint: convertchoices.Smith, Test: convertchoices.Jest, Build: convertchoices.Smith, CI: convertchoices.Smith}

func oldTooling() map[string]string {
	return map[string]string{
		"package.json": `{
  "scripts": {"build": "esbuild src/index.ts", "test": "jest"},
  "devDependencies": {"eslint": "^10", "@mnci/eslint-config": "*", "typescript-eslint": "^8", "nx": "1", "rollup": "^4", "jest": "^30", "vitest": "^3"}
}
`,
		"eslint.config.mjs":            sharedConfig,
		"packages/a/eslint.config.mjs": "export default []",
		"packages/a/rollup.config.cjs": "module.exports = {}",
		"vitest.config.ts":             "export default {}",
		"packages/a/package.json":      `{"name":"a","devDependencies":{"@nx/js":"1"}}`,
	}
}

func group(t *testing.T, p Plan, id string) Group {
	t.Helper()
	for _, g := range p.Groups {
		if g.ID == id {
			return g
		}
	}
	t.Fatalf("no %q group in %+v", id, p.Groups)
	return Group{}
}

func TestBuildMakesOneGroupPerRetiredTool(t *testing.T) {
	p := Build(workspace(oldTooling()), fullChoices, nil)
	var ids []string
	for _, g := range p.Groups {
		ids = append(ids, g.ID)
		if !g.Fix.Ask || g.Fix.Safe {
			t.Errorf("%s: a deletion is always asked, got ask=%v safe=%v", g.ID, g.Fix.Ask, g.Fix.Safe)
		}
	}
	if got := strings.Join(ids, ","); got != "lint,tests,build,nx" {
		t.Fatalf("groups = %s", got)
	}
	lint := group(t, p, GroupLint)
	if !strings.Contains(lint.Reason, "you chose smith's ESLint") || !strings.Contains(lint.Reason, "rules you wrote") {
		t.Errorf("reason = %q", lint.Reason)
	}
	if got := strings.Join(lint.Removes, ","); got != "@mnci/eslint-config,typescript-eslint" {
		t.Errorf("lint removes %q", got)
	}
	if tests := group(t, p, GroupTests); strings.Join(tests.Removes, ",") != "vitest" || tests.Files != 1 {
		t.Errorf("tests group = %+v", tests)
	}
	if nx := group(t, p, GroupNx); strings.Join(nx.Removes, ",") != "@nx/js,nx" {
		t.Errorf("nx removes %v", nx.Removes)
	}
}

func TestBuildAddsSmithsLintConfigToTheRootManifest(t *testing.T) {
	lint := group(t, Build(workspace(oldTooling()), fullChoices, nil), GroupLint)
	for _, e := range lint.Fix.Files {
		if e.Path == "package.json" {
			if !strings.Contains(e.Content, `"@monosmith/eslint-config": "^1.0.0"`) || strings.Contains(e.Content, "@mnci/eslint-config") {
				t.Fatalf("root manifest:\n%s", e.Content)
			}
			return
		}
	}
	t.Fatal("the root package.json is not edited")
}

func TestBuildSkipsADeclinedGroupAndPlansTheRestWithoutIt(t *testing.T) {
	p := Build(workspace(oldTooling()), fullChoices, map[string]bool{GroupLint: true})
	for _, g := range p.Groups {
		if g.ID == GroupLint {
			t.Fatal("the declined group is offered again")
		}
	}
	// The root lint config stays, so typescript-eslint is not touched by any group.
	for _, g := range p.Groups {
		for _, name := range g.Removes {
			if name == "typescript-eslint" {
				t.Fatalf("%s removes typescript-eslint though the lint config stays", g.ID)
			}
		}
	}
}

func TestBuildChangesNothingBelowTheFullLevel(t *testing.T) {
	c := fullChoices
	c.Level = convertchoices.LevelConservative
	if p := Build(workspace(oldTooling()), c, nil); len(p.Groups) != 0 {
		t.Fatalf("groups = %+v", p.Groups)
	}
}

func TestBuildKeepChoicesRemoveNothingOfThatTool(t *testing.T) {
	c := fullChoices
	c.Lint, c.Test, c.Build = convertchoices.Keep, convertchoices.Keep, convertchoices.Keep
	p := Build(workspace(oldTooling()), c, nil)
	if len(p.Groups) != 1 || p.Groups[0].ID != GroupNx {
		t.Fatalf("only the Nx group is left, got %+v", p.Groups)
	}
}

func TestBuildExplainsWhatAKeptConfigNeeds(t *testing.T) {
	files := oldTooling()
	files["package.json"] = strings.Replace(files["package.json"], `"build": "esbuild src/index.ts"`, `"build": "rollup -c"`, 1)
	p := Build(workspace(files), fullChoices, nil)
	build := Build(workspace(files), fullChoices, nil)
	for _, g := range build.Groups {
		if g.ID == GroupBuild {
			t.Fatalf("rollup is still run by a script, yet the build group removes %+v", g.Removes)
		}
	}
	found := false
	for _, k := range p.Kept {
		found = found || (k.What == "packages/a/rollup.config.cjs" && strings.Contains(k.Why, "rollup"))
	}
	if !found {
		t.Fatalf("kept = %+v", p.Kept)
	}
}

func TestTheConservativeFinishBuildsAndTestsButDoesNotLint(t *testing.T) {
	f := FinishBuildAndTest("smith")
	var got []string
	for _, s := range f.Steps[1:] {
		got = append(got, s.Description)
	}
	if strings.Join(got, "|") != StepBuild+"|"+StepTest+"|"+StepDoctor {
		t.Fatalf("checks = %v", got)
	}
}

func TestFinishInstallsOnceThenChecksTheWorkspace(t *testing.T) {
	f := Finish("smith", true)
	if f.Ask || !f.Safe || len(f.Steps) != 5 || f.Steps[0].Argv[1] != "npm-lockfile" || f.Steps[4].Argv[1] != "doctor" {
		t.Fatalf("steps = %+v", f.Steps)
	}
	if got := Finish("smith", false); len(got.Steps) != 1 {
		t.Fatalf("without the checks only the install remains, got %+v", got.Steps)
	}
}
