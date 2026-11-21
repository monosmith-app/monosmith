package projecttooling

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// typeScript4Fixture is a workspace whose root holds the TypeScript 4 that
// adopt hoisted from the project.
var typeScript4Fixture = map[string]string{
	"package.json":               `{"name": "w", "private": true, "workspaces": ["packages/*"], "devDependencies": {"typescript": "^4.9.5"}}`,
	"packages/mitt/package.json": `{"name": "mitt", "scripts": {"typecheck": "tsc --noEmit"}, "devDependencies": {"typescript": "~4.4.0"}}`,
}

func TestPlanTypeScriptRaisesATypeScript4InTheRootAndTheProject(t *testing.T) {
	root := writeTree(t, typeScript4Fixture)
	planned, err := PlanTypeScript(root, "packages/mitt")
	if err != nil || planned == nil {
		t.Fatalf("plan %v, %v", planned, err)
	}
	if planned.Ask || !planned.Safe || len(planned.Files) != 2 {
		t.Fatalf("fix %+v", planned.Fix)
	}
	for _, f := range planned.Files {
		if !strings.Contains(f.Content, `"typescript": "`+TypeScriptRange+`"`) {
			t.Errorf("%s:\n%s", f.Path, f.Content)
		}
	}
	if !strings.Contains(planned.Description, "from ^4.9.5 to "+TypeScriptRange) {
		t.Errorf("description %q", planned.Description)
	}
}

func TestPlanTypeScriptLeavesATypeScript5OrARangeItCannotRead(t *testing.T) {
	root := writeTree(t, map[string]string{
		"package.json":            `{"name": "w", "devDependencies": {"typescript": "^5.2.0"}}`,
		"packages/a/package.json": `{"name": "a", "devDependencies": {"typescript": "next"}}`,
	})
	if planned, err := PlanTypeScript(root, "packages/a"); err != nil || planned != nil {
		t.Errorf("plan %+v, %v", planned, err)
	}
}

func TestRaiseTypeScriptWithYesAppliesIt(t *testing.T) {
	root := writeTree(t, typeScript4Fixture)
	var out bytes.Buffer
	left, err := raiseTypeScript(context.Background(), Input{Root: root, Dir: "packages/mitt"}, "mitt", RunOptions{Out: &out, Yes: true})
	if err != nil || len(left) != 0 {
		t.Fatalf("left %v, %v\n%s", left, err, out.String())
	}
	if !strings.Contains(readFile(t, root, "package.json"), TypeScriptRange) {
		t.Errorf("root package.json:\n%s", readFile(t, root, "package.json"))
	}
}

func TestRaiseTypeScriptWithoutAnAnswerLeavesTheStep(t *testing.T) {
	root := writeTree(t, typeScript4Fixture)
	left, err := raiseTypeScript(context.Background(), Input{Root: root, Dir: "packages/mitt"}, "mitt", RunOptions{Out: &bytes.Buffer{}})
	if err != nil || len(left) != 1 || !strings.Contains(left[0], "Set typescript to "+TypeScriptRange) {
		t.Fatalf("left %v, %v", left, err)
	}
	if strings.Contains(readFile(t, root, "package.json"), TypeScriptRange) {
		t.Error("the raise was applied without an answer")
	}
}
