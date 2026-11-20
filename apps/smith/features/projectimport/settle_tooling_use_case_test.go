package projectimport

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// legacySource is a create-react-app project with an eslintConfig key.
func legacySource(t *testing.T) Source {
	t.Helper()
	dir := t.TempDir()
	manifest := `{"name": "demo", "scripts": {"build": "react-scripts build"}, "dependencies": {"react-scripts": "4.0.3"}, "eslintConfig": {"extends": ["react-app"]}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	return Source{Dir: dir}
}

func TestSettleToolingTakesTheFlagAndPutsTheOwnESLintSetupAside(t *testing.T) {
	p := Plan{Name: "demo"}
	o := Options{Tooling: []string{"smith"}, Out: &bytes.Buffer{}}
	if err := SettleTooling(legacySource(t), &p, &o); err != nil {
		t.Fatal(err)
	}
	if p.Tooling != "smith" || o.ESLint != ESLintShared {
		t.Errorf("tooling %q, eslint setup %q", p.Tooling, o.ESLint)
	}
	if len(p.Legacy) == 0 || !strings.Contains(p.Legacy[0], "create-react-app") {
		t.Errorf("legacy %q", p.Legacy)
	}
	if !strings.Contains(Summary(p), "smith's build and ESLint config go in") {
		t.Errorf("summary:\n%s", Summary(p))
	}
}

func TestSettleToolingWithYesKeepsTheToolsAndSaysHowToChooseSmiths(t *testing.T) {
	var out bytes.Buffer
	p := Plan{Name: "demo"}
	o := Options{Yes: true, Out: &out}
	if err := SettleTooling(legacySource(t), &p, &o); err != nil {
		t.Fatal(err)
	}
	if p.Tooling != "keep" || o.ESLint != "" {
		t.Errorf("tooling %q, eslint setup %q", p.Tooling, o.ESLint)
	}
	if !strings.Contains(out.String(), "--tooling demo=smith") {
		t.Errorf("output %q", out.String())
	}
}

func TestSettleToolingAsksNothingOfAProjectInAnotherLanguage(t *testing.T) {
	p := Plan{Name: "lib"}
	o := Options{Out: &bytes.Buffer{}}
	if err := SettleTooling(Source{Dir: t.TempDir()}, &p, &o); err != nil {
		t.Fatal(err)
	}
	if p.Tooling != "" || !strings.Contains(Summary(p), "Node projects only") {
		t.Errorf("tooling %q:\n%s", p.Tooling, Summary(p))
	}
}
