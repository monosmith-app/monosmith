package lockcarry

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

const workspaceLock = `{
  "name": "ws",
  "lockfileVersion": 3,
  "requires": true,
  "packages": {
    "": {"name": "ws", "workspaces": ["apps/*"], "devDependencies": {"eslint": ">=9"}},
    "node_modules/eslint": {"version": "9.0.0"}
  }
}`

const manifest = `{"name": "web", "dependencies": {"react-icons": "^4.3.1", "local": "file:../local"}, "devDependencies": {"jest": "^27.0.0"}}`

func packagesOf(t *testing.T, raw []byte) map[string]json.RawMessage {
	t.Helper()
	var l npmLock
	if err := json.Unmarshal(raw, &l); err != nil {
		t.Fatal(err)
	}
	return l.Packages
}

func TestMergeNpmAddsTheProjectEntriesUnderItsFolder(t *testing.T) {
	source := `{
  "lockfileVersion": 2,
  "packages": {
    "": {"name": "web"},
    "node_modules/react-icons": {"version": "4.3.1", "resolved": "https://r/react-icons-4.3.1.tgz"},
    "node_modules/jest": {"version": "27.3.1", "dev": true},
    "node_modules/jest/node_modules/ms": {"version": "2.0.0"},
    "node_modules/local": {"resolved": "../local", "link": true},
    "node_modules/local2": {"version": "file:../local2"}
  }
}`
	out, kept, added, err := MergeNpm([]byte(workspaceLock), []byte(source), []byte(manifest), "apps/web")
	if err != nil {
		t.Fatal(err)
	}
	if added != 3 {
		t.Errorf("added %d entries, want 3 (a link and a local folder stay out)", added)
	}
	pkgs := packagesOf(t, out)
	for _, key := range []string{"apps/web/node_modules/react-icons", "apps/web/node_modules/jest", "apps/web/node_modules/jest/node_modules/ms", "node_modules/eslint", ""} {
		if _, ok := pkgs[key]; !ok {
			t.Errorf("%q is missing from the merged lockfile", key)
		}
	}
	if _, ok := pkgs["apps/web/node_modules/local"]; ok {
		t.Error("a link of the source workspace was copied")
	}
	want := []Locked{{Name: "jest", Version: "27.3.1"}, {Name: "react-icons", Version: "4.3.1"}}
	if !reflect.DeepEqual(kept, want) {
		t.Errorf("kept = %v, want %v", kept, want)
	}
	if !strings.Contains(string(out), `">=9"`) {
		t.Error("the merge escapes > in a range, which npm writes as is")
	}
	if !strings.HasPrefix(string(out), "{\n  \"name\": \"ws\",\n  \"lockfileVersion\": 3,") {
		t.Errorf("the top keys lost npm's order:\n%.80s", out)
	}
}

func TestMergeNpmKeepsAnEntryTheWorkspaceHasAlready(t *testing.T) {
	ws := `{"lockfileVersion": 3, "packages": {"": {}, "apps/web/node_modules/jest": {"version": "29.0.0"}}}`
	source := `{"lockfileVersion": 3, "packages": {"node_modules/jest": {"version": "27.3.1"}}}`
	out, _, added, err := MergeNpm([]byte(ws), []byte(source), []byte(manifest), "apps/web/")
	if err != nil || added != 0 {
		t.Fatalf("added %d, %v", added, err)
	}
	if !strings.Contains(string(packagesOf(t, out)["apps/web/node_modules/jest"]), "29.0.0") {
		t.Error("the workspace's own entry was replaced")
	}
}

func TestMergeNpmReadsTheNestedTreeOfVersion1(t *testing.T) {
	source := `{"lockfileVersion": 1, "dependencies": {
  "react-icons": {"version": "4.3.1", "resolved": "https://r/x.tgz", "integrity": "sha512-x", "requires": {"tslib": "^2"},
    "dependencies": {"tslib": {"version": "2.0.0", "dev": true}}}
}}`
	out, kept, added, err := MergeNpm([]byte(workspaceLock), []byte(source), []byte(manifest), "apps/web")
	if err != nil || added != 2 {
		t.Fatalf("added %d, %v", added, err)
	}
	var e struct {
		Version      string            `json:"version"`
		Integrity    string            `json:"integrity"`
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(packagesOf(t, out)["apps/web/node_modules/react-icons"], &e); err != nil {
		t.Fatal(err)
	}
	if e.Version != "4.3.1" || e.Integrity != "sha512-x" || e.Dependencies["tslib"] != "^2" {
		t.Errorf("entry = %+v", e)
	}
	if _, ok := packagesOf(t, out)["apps/web/node_modules/react-icons/node_modules/tslib"]; !ok {
		t.Error("a nested dependency of version 1 is missing")
	}
	if !reflect.DeepEqual(kept, []Locked{{Name: "react-icons", Version: "4.3.1"}}) {
		t.Errorf("kept = %v", kept)
	}
}

func TestMergeNpmRefusesAWorkspaceLockWithoutPackages(t *testing.T) {
	if _, _, _, err := MergeNpm([]byte(`{"lockfileVersion": 1}`), []byte(`{}`), []byte(manifest), "apps/web"); err != ErrNoPackages {
		t.Errorf("err = %v, want ErrNoPackages", err)
	}
}
