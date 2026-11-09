package upgrade

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestAdoptDefaultsRecordsTheLatestVintageAndKeepsComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, workspace.FileName)
	src := "schema = 1\n\n[workspace]\nname = \"w\" # the name\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	change, err := AdoptDefaults(dir, true)
	if err != nil || !change.Changed() || change.From != "" {
		t.Fatalf("check: %+v %v", change, err)
	}
	if raw, _ := os.ReadFile(path); string(raw) != src {
		t.Fatalf("check wrote the file: %q", raw)
	}
	if _, err := AdoptDefaults(dir, false); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	if !strings.Contains(string(raw), "# the name") || !strings.Contains(string(raw), `defaults = "`+workspace.LatestDefaultsVintage()+`"`) {
		t.Fatalf("not recorded: %q", raw)
	}
	if change, _ := AdoptDefaults(dir, false); change.Changed() {
		t.Errorf("second run still changes: %+v", change)
	}
}

func TestAdoptDefaultsRefusesAVintageNewerThanSmith(t *testing.T) {
	dir := t.TempDir()
	src := "[workspace]\nname = \"w\"\ndefaults = \"2999.01\"\n"
	if err := os.WriteFile(filepath.Join(dir, workspace.FileName), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := AdoptDefaults(dir, false); err == nil || !strings.Contains(err.Error(), "upgrade smith") {
		t.Fatalf("got %v", err)
	}
}
