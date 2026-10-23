package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var nodeProject = []protocol.Project{{Name: "api", Language: "node"}}

func configIn(t *testing.T, text string) string {
	t.Helper()
	dir := t.TempDir()
	if text != "" {
		if err := os.WriteFile(filepath.Join(dir, "eslint.config.mjs"), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestAConfigThatLoadsMonosmithPasses(t *testing.T) {
	dir := configIn(t, "import monosmith from '@monosmith/eslint-config'\nexport default monosmith({})\n")
	if f := checkNodeESLint(dir, nodeProject); len(f) != 1 || !f[0].OK {
		t.Fatalf("%+v", f)
	}
}

func TestAnMnciConfigFailsAndNamesWhatItUses(t *testing.T) {
	dir := configIn(t, "import mnci from '@mnci/eslint-config'\n")
	f := checkNodeESLint(dir, nodeProject)
	if len(f) != 1 || f[0].OK || !strings.Contains(f[0].Detail, "@mnci/eslint-config") || !strings.Contains(f[0].Remedy, "@monosmith/eslint-config") {
		t.Fatalf("%+v", f)
	}
}

func TestNoConfigFails(t *testing.T) {
	if f := checkNodeESLint(configIn(t, ""), nodeProject); len(f) != 1 || f[0].OK {
		t.Fatalf("%+v", f)
	}
}

func TestOptingOutOfSlicesIsRespected(t *testing.T) {
	dir := configIn(t, "import m from '@monosmith/eslint-config'\nexport default m({ verticalSlices: false })\n")
	if f := checkNodeESLint(dir, nodeProject); f != nil {
		t.Fatalf("%+v", f)
	}
}

func TestOnlyNodeProjectsAreChecked(t *testing.T) {
	if f := checkNodeESLint(configIn(t, ""), []protocol.Project{{Name: "tool", Language: "go"}}); f != nil {
		t.Fatalf("%+v", f)
	}
}
