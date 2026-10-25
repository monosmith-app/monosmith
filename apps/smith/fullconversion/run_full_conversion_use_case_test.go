package fullconversion

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
)

// oldTooling makes a git repository with a root lint config built on the
// old shared one, a build config nothing runs and a manifest naming both.
func oldTooling(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	files := map[string]string{
		"package.json": `{
  "scripts": {"build": "esbuild src/index.ts"},
  "devDependencies": {"eslint": "^10", "@mnci/eslint-config": "*", "rollup": "^4"}
}
`,
		"eslint.config.mjs":            "import mnci from '@mnci/eslint-config'\n\nexport default [...mnci({ workspaceRoot: import.meta.dirname })]\n",
		"packages/a/rollup.config.cjs": "module.exports = {}\n",
	}
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := exec.Command("git", "-C", dir, "init", "-q").Run(); err != nil {
		t.Fatal(err)
	}
	return dir
}

func options(dir string, answers string, out *bytes.Buffer) RunOptions {
	return RunOptions{
		Root: dir, Out: out, In: bufio.NewReader(strings.NewReader(answers)), Interactive: true,
		Choices:    convertchoices.Choices{Level: convertchoices.LevelFull, Lint: convertchoices.Smith, Test: convertchoices.Keep, Build: convertchoices.Smith, CI: convertchoices.Smith},
		SkipFinish: true, Smith: "smith", Files: gitclient.New(dir).Files,
	}
}

func read(t *testing.T, dir, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
	if err != nil {
		return ""
	}
	return string(raw)
}

func TestRunAppliesTheGroupsAnsweredYesAndLeavesTheOthers(t *testing.T) {
	dir := oldTooling(t)
	var out bytes.Buffer
	// Lint: yes. Build: no.
	res, err := Run(context.Background(), options(dir, "y\nn\n", &out))
	if err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	if strings.Join(res.Applied, ",") != "lint" || strings.Join(res.Declined, ",") != "build" {
		t.Fatalf("applied %v declined %v\n%s", res.Applied, res.Declined, out.String())
	}
	if cfg := read(t, dir, "eslint.config.mjs"); !strings.Contains(cfg, "@monosmith/eslint-config") || strings.Contains(cfg, "mnci") {
		t.Errorf("the lint config was not moved over:\n%s", cfg)
	}
	manifest := read(t, dir, "package.json")
	if strings.Contains(manifest, "@mnci/eslint-config") || !strings.Contains(manifest, `"@monosmith/eslint-config"`) {
		t.Errorf("the lint packages were not swapped:\n%s", manifest)
	}
	if !strings.Contains(manifest, `"rollup"`) || read(t, dir, "packages/a/rollup.config.cjs") == "" {
		t.Error("the declined build group removed something")
	}
}

func TestRunWithYesAppliesEverythingWithoutAsking(t *testing.T) {
	dir := oldTooling(t)
	var out bytes.Buffer
	o := options(dir, "", &out)
	o.Interactive, o.In, o.Yes = false, nil, true
	res, err := Run(context.Background(), o)
	if err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	if strings.Join(res.Applied, ",") != "lint,build" {
		t.Fatalf("applied %v\n%s", res.Applied, out.String())
	}
	if read(t, dir, "packages/a/rollup.config.cjs") != "" || strings.Contains(read(t, dir, "package.json"), "rollup") {
		t.Error("the build group was not applied")
	}
}

func TestRunPlansTheNextGroupFromTheFilesAsTheyAre(t *testing.T) {
	dir := oldTooling(t)
	var out bytes.Buffer
	// Lint declined: the root config still imports the old shared config, so
	// its package must stay when the other groups are planned.
	res, err := Run(context.Background(), options(dir, "n\ny\n", &out))
	if err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	if strings.Join(res.Declined, ",") != "lint" || strings.Join(res.Applied, ",") != "build" {
		t.Fatalf("applied %v declined %v", res.Applied, res.Declined)
	}
	if !strings.Contains(read(t, dir, "package.json"), "@mnci/eslint-config") {
		t.Error("a declined lint group still lost its package")
	}
}
