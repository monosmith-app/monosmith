package toolcatalog

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

var knownManagers = []string{"winget", "brew", "brew-cask", "apt", "dnf", "pacman", "npm", "uv", "fnm", "rustup"}

func TestEntriesAreComplete(t *testing.T) {
	date := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	seen := map[string]bool{}
	for _, e := range entries {
		if seen[e.Name] {
			t.Errorf("%s: listed twice", e.Name)
		}
		seen[e.Name] = true
		if e.Title == "" || e.Why == "" || e.Consequence == "" || e.Program == "" {
			t.Errorf("%s: title, why, consequence and program are required", e.Name)
		}
		if !strings.HasPrefix(e.Download, "https://") || !strings.HasPrefix(e.Source, "https://") {
			t.Errorf("%s: download and source must be https addresses", e.Name)
		}
		if !date.MatchString(e.Checked) {
			t.Errorf("%s: checked must be a date, got %q", e.Name, e.Checked)
		}
		for _, in := range e.Install {
			if !slices.Contains(knownManagers, in.Manager) || in.Package == "" {
				t.Errorf("%s: bad install entry %+v", e.Name, in)
			}
			if strings.Contains(in.Package, "{") && e.Suggested == "" {
				t.Errorf("%s: %q needs a suggested version", e.Name, in.Package)
			}
		}
	}
}

func TestForPanicsOnUnknownName(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("want a panic for an unknown tool")
		}
	}()
	For("nope")
}
