package adoption

import (
	"strings"
	"testing"
)

func TestExistingWorkflowsWarnsAboutEachWorkflowOfTheUsers(t *testing.T) {
	files := []string{".github/workflows/mine.yml", ".github/workflows/ci.yml", ".github/workflows/own.yaml", ".github/workflows/readme.md", "src/a.ts"}
	read := func(f string) string {
		if f == ".github/workflows/own.yaml" {
			return "# " + GeneratedMarker + "\n"
		}
		return "on: push\n"
	}
	items := ExistingWorkflows(files, read, []string{"github"})
	if len(items) != 1 || !strings.Contains(items[0].Message, ".github/workflows/mine.yml") || !strings.Contains(items[0].Message, "CI would run twice") {
		t.Fatalf("items = %+v", items)
	}
	if len(items[0].Steps) < 2 || items[0].Step == "" {
		t.Errorf("each way out is named: %+v", items[0])
	}
	if got := ExistingWorkflows(files, read, []string{"azure"}); len(got) != 0 {
		t.Errorf("a workspace that does not use GitHub has nothing to run twice: %+v", got)
	}
}

func TestCommitWithoutInitIsRefusedBeforeAnythingIsRead(t *testing.T) {
	cmd := Command("test")
	cmd.SetArgs([]string{"--commit"})
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--init") {
		t.Fatalf("err = %v", err)
	}
}

func TestPnpmWorkspaceIsRecognisedByItsFileOrItsRanges(t *testing.T) {
	cases := []struct {
		name  string
		files []string
		text  map[string]string
		want  bool
	}{
		{"pnpm-workspace.yaml", []string{"pnpm-workspace.yaml", "package.json"}, nil, true},
		{"workspace range", []string{"package.json", "packages/a/package.json"}, map[string]string{"packages/a/package.json": `{"dependencies": {"b": "workspace:*"}}`}, true},
		{"catalog range", []string{"package.json"}, map[string]string{"package.json": `{"devDependencies": {"react": "catalog:react"}}`}, true},
		{"plain npm", []string{"package.json"}, map[string]string{"package.json": `{"dependencies": {"b": "^1.0.0"}}`}, false},
	}
	for _, c := range cases {
		item, got := PnpmWorkspace(c.files, func(f string) string { return c.text[f] })
		if got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
		if got && (item.Blocks != BlocksAdoption || !strings.Contains(item.Message, "pnpm workspace") || len(item.Steps) == 0) {
			t.Errorf("%s: the blocker must stop --init and name the way on: %+v", c.name, item)
		}
	}
}
