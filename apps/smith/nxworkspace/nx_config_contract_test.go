package nxworkspace

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

func TestNxJSONReadsTheNx24ListFormOfTargetDefaults(t *testing.T) {
	raw := `{"targetDefaults": {
		"build": {"cache": true, "outputs": ["dist"]},
		"test": [{"cache": true}, {"projects": ["a"], "outputs": ["coverage"]}]
	}}`
	var nx NxJSON
	if err := json.Unmarshal([]byte(raw), &nx); err != nil {
		t.Fatal(err)
	}
	if got := nx.TargetDefaults["build"].Outputs; len(got) != 1 || got[0] != "dist" {
		t.Errorf("the object form is read as before: %v", nx.TargetDefaults["build"])
	}
	test, ok := nx.TargetDefaults["test"]
	if !ok || test.Cache == nil || !*test.Cache || len(test.Outputs) != 1 {
		t.Errorf("the entries of the list are merged in order: %+v", test)
	}
}

func TestNxJSONReadsTheMnciCIAsAStringOrAList(t *testing.T) {
	for raw, want := range map[string]int{`{"mnci": {"ci": "github"}}`: 1, `{"mnci": {"ci": ["github", "azure"]}}`: 2} {
		var nx NxJSON
		if err := json.Unmarshal([]byte(raw), &nx); err != nil {
			t.Fatalf("%s: %v", raw, err)
		}
		if len(nx.Mnci.CI) != want {
			t.Errorf("%s: ci = %v", raw, nx.Mnci.CI)
		}
	}
}

func TestReadJSONNamesTheFieldNotTheGoType(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "nx.json"), []byte(`{"defaultBase": ["main"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var nx NxJSON
	err := readJSON(dir, "nx.json", &nx)
	var p *problem.Error
	if !errors.As(err, &p) {
		t.Fatalf("want a problem with a way on, got %v", err)
	}
	for _, bad := range []string{"NxJSON", "Go struct", "nxworkspace."} {
		if strings.Contains(err.Error(), bad) {
			t.Errorf("the message names a Go type (%q): %v", bad, err)
		}
	}
	if !strings.Contains(err.Error(), "defaultBase") {
		t.Errorf("the message does not say which field: %v", err)
	}
}
