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

func TestThePipelineSmithMustMatchTheDependency(t *testing.T) {
	pkg := `{"devDependencies": {"monosmith": "^1.90.0"}}` + "\n"
	lock := `{"packages": {"node_modules/monosmith": {"version": "1.95.0"}}}`

	root := workspaceWith(t, map[string]string{"package.json": pkg, "package-lock.json": lock})
	got := smithDependencyFinding(root, "1.100.0")
	if len(got) != 1 || got[0].OK || !got[0].Warning || got[0].Fix == nil {
		t.Fatalf("a locked 1.95.0 against a 1.100.0 pipeline warns with a fix: %+v", got)
	}
	if edit := got[0].Fix.Files[0]; edit.Path != "package.json" || !strings.Contains(edit.Content, `"monosmith": "1.100.0"`) {
		t.Errorf("the fix sets the pipeline's version: %+v", edit)
	}
	if got := smithDependencyFinding(root, "1.95.0"); got != nil {
		t.Errorf("matching versions are silent: %+v", got)
	}
	if got := smithDependencyFinding(root, "latest"); got != nil {
		t.Errorf("latest is silent: %+v", got)
	}
}

func TestTheDeclaredVersionCountsOnlyWhenExactAndUnlocked(t *testing.T) {
	exact := workspaceWith(t, map[string]string{"package.json": `{"devDependencies": {"monosmith": "1.90.0"}}`})
	if got := smithDependencyFinding(exact, "1.100.0"); len(got) != 1 {
		t.Errorf("an exact declared version is compared: %+v", got)
	}
	ranged := workspaceWith(t, map[string]string{"package.json": `{"devDependencies": {"monosmith": "^1.90.0"}}`})
	if got := smithDependencyFinding(ranged, "1.100.0"); got != nil {
		t.Errorf("a range with no lock cannot be compared: %+v", got)
	}
	none := workspaceWith(t, map[string]string{"package.json": `{"devDependencies": {"jest": "30.0.0"}}`})
	if got := smithDependencyFinding(none, "1.100.0"); got != nil {
		t.Errorf("no monosmith dependency, no finding: %+v", got)
	}
	if got := smithDependencyFinding(t.TempDir(), "1.100.0"); got != nil {
		t.Errorf("no package.json, no finding: %+v", got)
	}
}
