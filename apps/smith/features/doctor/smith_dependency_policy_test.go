package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func workspaceWith(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestAnOldMonosmithDependencyMovesToTheLauncher(t *testing.T) {
	pkg := `{"devDependencies": {"monosmith": "^2.40.0"}}` + "\n"
	lock := `{"packages": {"node_modules/monosmith": {"version": "2.48.0"}}}`
	toml := "[workspace]\nname = \"acme\"\n"

	root := workspaceWith(t, map[string]string{"package.json": pkg, "package-lock.json": lock, "smith.toml": toml})
	got := smithDependencyFinding(root, "")
	if len(got) != 1 || got[0].OK || !got[0].Warning || got[0].Fix == nil {
		t.Fatalf("a locked 2.48.0 is the old binary package: a warning with a fix, got %+v", got)
	}
	files := got[0].Fix.Files
	if len(files) != 2 || !strings.Contains(files[0].Content, `"monosmith": "^3"`) {
		t.Fatalf("the fix depends on the launcher: %+v", files)
	}
	if files[1].Path != "smith.toml" || !strings.Contains(files[1].Content, `smith = "2.48.0"`) {
		t.Errorf("without a pin, the version it ran becomes the pin: %+v", files[1])
	}
	steps := got[0].Fix.Steps
	if len(steps) != 1 || strings.Join(steps[0].Argv, " ") != "npm install --no-audit --no-fund" || steps[0].Dir != "" {
		t.Errorf("the fix then runs npm install at the workspace root to refresh the lockfile: %+v", steps)
	}
	if pinned := smithDependencyFinding(root, "2.50.0"); len(pinned) != 1 || len(pinned[0].Fix.Files) != 1 {
		t.Errorf("a workspace that pins keeps its pin: %+v", pinned)
	}
	launcher := workspaceWith(t, map[string]string{"package.json": `{"devDependencies": {"monosmith": "^3"}}`, "package-lock.json": `{"packages": {"node_modules/monosmith": {"version": "3.0.0"}}}`})
	if got := smithDependencyFinding(launcher, ""); got != nil {
		t.Errorf("the launcher is silent: %+v", got)
	}
}

func TestTheDeclaredVersionCountsOnlyWhenExactAndUnlocked(t *testing.T) {
	exact := workspaceWith(t, map[string]string{"package.json": `{"devDependencies": {"monosmith": "2.10.0"}}`})
	if got := smithDependencyFinding(exact, ""); len(got) != 1 {
		t.Errorf("an exact declared version is compared: %+v", got)
	}
	ranged := workspaceWith(t, map[string]string{"package.json": `{"devDependencies": {"monosmith": "^2.10.0"}}`})
	if got := smithDependencyFinding(ranged, ""); got != nil {
		t.Errorf("a range with no lock cannot be read: %+v", got)
	}
	none := workspaceWith(t, map[string]string{"package.json": `{"devDependencies": {"jest": "30.0.0"}}`})
	if got := smithDependencyFinding(none, ""); got != nil {
		t.Errorf("no monosmith dependency, no finding: %+v", got)
	}
	if got := smithDependencyFinding(t.TempDir(), ""); got != nil {
		t.Errorf("no package.json, no finding: %+v", got)
	}
}
