//go:build e2e

package goldenpath

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// deriveLibs are the packages of the workspace and the ones each depends on.
var deriveLibs = []struct {
	name string
	deps []string
}{
	{"l0", nil}, {"l1", []string{"l0"}}, {"l2", []string{"l0", "l1"}}, {"l3", []string{"l2"}}, {"l4", []string{"l0", "l3"}}, {"l5", []string{"l4"}},
}

// jestMarkers is a Jest config whose displayName sits between the markers; the
// second %s holds the project's own settings.
const jestMarkers = "module.exports = {\n  // smith:begin node.jest.display-name\n%s  // smith:end node.jest.display-name\n%s}\n"

// TestDeriveRegeneratesTheRepeatedConfigOfAConvertedWorkspace is the exit test
// of derived files (ADR 0046): in a workspace like a converted one, whose
// packages repeat a tsconfig.lib.json that differs only by its references and
// a Jest displayName, the repeated parts are deleted, smith derive writes them
// back as they were, build and test give the same result, a hand edit is
// reported as drift and never overwritten, and derive --check fails and then
// passes.
func TestDeriveRegeneratesTheRepeatedConfigOfAConvertedWorkspace(t *testing.T) {
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm is not installed")
	}
	base := t.TempDir()
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: identity()}
	w.smith(true, "new", "dw", "--ci", "github")
	w.dir = filepath.Join(base, "dw")
	for _, l := range deriveLibs {
		w.smith(true, "add", "lib", l.name, "--lang", "node")
	}
	for _, l := range deriveLibs {
		manifest := fmt.Sprintf("libs/%s/package.json", l.name)
		if len(l.deps) > 0 {
			var deps []string
			for _, d := range l.deps {
				deps = append(deps, fmt.Sprintf("%q: %q", "@dw/"+d, "*"))
			}
			w.write(manifest, strings.Replace(w.read(manifest), `"scripts"`, `"dependencies": { `+strings.Join(deps, ", ")+` },`+"\n  "+`"scripts"`, 1))
		}
		// The repeated config, as a converted workspace has it.
		var refs []string
		for _, d := range l.deps {
			refs = append(refs, fmt.Sprintf(`{ "path": "../%s/tsconfig.lib.json" }`, d))
		}
		w.write(fmt.Sprintf("libs/%s/tsconfig.lib.json", l.name), `{ "references": [`+strings.Join(refs, ", ")+"] }\n")
		moveJestConfig(t, w, l.name)
	}
	w.write("smith.toml", w.read("smith.toml")+"\n[derive]\nenabled = true\n")
	w.smith(true, "install")

	// Derived files are opt-in: without the setting the command says so.
	// (Here it is on, so the first run records the files that are right.)
	first := w.smith(true, "derive")
	if !strings.Contains(first, "recorded") {
		t.Fatalf("the first run records the files that already match:\n%s", first)
	}
	w.smith(true, "derive", "--check")
	if !strings.Contains(w.read("smith.lock"), "owner = 'smith'") {
		t.Fatalf("smith.lock does not record the files:\n%s", w.read("smith.lock"))
	}
	files := derivedFiles()
	before := map[string]any{}
	for _, f := range files {
		if strings.HasSuffix(f, ".json") {
			before[f] = parseJSON(t, w.read(f))
		}
	}
	healthy := func() string {
		out := w.smith(true, "run", "build", "test")
		return regexp.MustCompile(`\d+/\d+ tasks succeeded`).FindString(out)
	}
	wantResult := healthy()
	if wantResult == "" {
		t.Fatal("smith run build test reports no result")
	}
	w.git("init", "-q")
	w.git("add", "-A")
	w.git("commit", "-qm", "chore: a converted workspace")

	// Delete the repeated files and the derived lines.
	for _, l := range deriveLibs {
		if err := os.Remove(filepath.Join(w.dir, "libs", l.name, "tsconfig.lib.json")); err != nil {
			t.Fatal(err)
		}
		w.write(fmt.Sprintf("libs/%s/jest.config.cjs", l.name), fmt.Sprintf(jestMarkers, "", jestRest[l.name]))
	}
	if out := w.smith(false, "derive", "--check"); !strings.Contains(out, "out of date") {
		t.Fatalf("--check does not fail on missing derived files:\n%s", out)
	}
	w.smith(true, "derive")
	w.smith(true, "derive", "--check")
	for _, f := range files {
		if strings.HasSuffix(f, ".json") {
			if got := parseJSON(t, w.read(f)); !reflect.DeepEqual(got, before[f]) {
				t.Errorf("%s was not regenerated as it was:\n%v\nwant\n%v", f, got, before[f])
			}
		}
	}
	if got := w.read("libs/l4/jest.config.cjs"); !strings.Contains(got, "displayName: '@dw/l4',") {
		t.Errorf("the display name was not regenerated:\n%s", got)
	}
	if got := healthy(); got != wantResult {
		t.Errorf("build and test: %q, before the deletion %q", got, wantResult)
	}

	// A hand edit is drift: reported, not overwritten, and --check fails.
	edited := `{ "extends": "../../tsconfig.json", "references": [{ "path": "../hand" }] }` + "\n"
	w.write("libs/l2/tsconfig.lib.json", edited)
	out := w.smith(false, "derive")
	if !strings.Contains(out, "drift") || !strings.Contains(out, "libs/l2/tsconfig.lib.json") || w.read("libs/l2/tsconfig.lib.json") != edited {
		t.Fatalf("a hand edit must be reported and kept:\n%s", out)
	}
	w.smith(false, "derive", "--check")
	if doctor := w.smith(false, "doctor"); !strings.Contains(doctor, "derived file libs/l2/tsconfig.lib.json") || !strings.Contains(doctor, "owner: smith") {
		t.Fatalf("doctor does not report the drift with its owner:\n%s", doctor)
	}
	if explain := w.smith(true, "derive", "--explain", "libs/l2/tsconfig.lib.json"); !strings.Contains(explain, "drift") || !strings.Contains(explain, "../hand") {
		t.Fatalf("explain:\n%s", explain)
	}

	// The choices: eject keeps the edit and stops touching the file; replace
	// takes the derived content back.
	w.smith(true, "derive", "--eject", "libs/l2/tsconfig.lib.json")
	w.smith(true, "derive", "--check")
	if w.read("libs/l2/tsconfig.lib.json") != edited {
		t.Fatal("eject changed the file")
	}
	w.write("libs/l3/tsconfig.lib.json", `{ "references": [{ "path": "../hand" }] }`+"\n")
	w.smith(false, "derive")
	w.smith(true, "derive", "--replace", "libs/l3/tsconfig.lib.json")
	w.smith(true, "derive", "--check")
	if got := parseJSON(t, w.read("libs/l3/tsconfig.lib.json")); !reflect.DeepEqual(got, before["libs/l3/tsconfig.lib.json"]) {
		t.Errorf("replace did not restore the derived content: %v", got)
	}
}

// jestRest holds, per package, the lines of its Jest config outside the markers.
var jestRest = map[string]string{}

// moveJestConfig turns the jest key of a package's manifest into a
// jest.config.cjs with the displayName between the markers, as a converted
// workspace has it.
func moveJestConfig(t *testing.T, w *workspace, name string) {
	t.Helper()
	manifest := fmt.Sprintf("libs/%s/package.json", name)
	var m map[string]any
	if err := json.Unmarshal([]byte(w.read(manifest)), &m); err != nil {
		t.Fatal(err)
	}
	rest, _ := m["jest"].(map[string]any)
	delete(m, "jest")
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	w.write(manifest, string(out)+"\n")
	var lines string
	if len(rest) > 0 {
		b, _ := json.MarshalIndent(rest, "  ", "  ")
		lines = strings.TrimSuffix(strings.TrimPrefix(string(b), "{\n"), "\n  }") + ",\n"
	}
	jestRest[name] = lines
	w.write(fmt.Sprintf("libs/%s/jest.config.cjs", name), fmt.Sprintf(jestMarkers, fmt.Sprintf("  displayName: '@dw/%s',\n", name), lines))
}

// derivedFiles are the files of the workspace the renderers own.
func derivedFiles() []string {
	var out []string
	for _, l := range deriveLibs {
		out = append(out, fmt.Sprintf("libs/%s/tsconfig.lib.json", l.name))
	}
	return out
}

func parseJSON(t *testing.T, s string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("%v:\n%s", err, s)
	}
	return v
}
