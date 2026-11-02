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

// TestEveryInstallEntryHasASourceAndADate fails for an install entry that no
// catalogue page covers or whose tool carries no check date.
func TestEveryInstallEntryHasASourceAndADate(t *testing.T) {
	for _, e := range entries {
		for _, in := range e.Install {
			if installCatalogues[in.Manager] == "" {
				t.Errorf("%s/%s: no catalogue page is recorded for this manager", e.Name, in.Manager)
			}
			if !strings.HasPrefix(e.Source, "https://") || e.Checked == "" {
				t.Errorf("%s/%s: the entry needs a source address and a check date", e.Name, in.Manager)
			}
		}
	}
	for manager, page := range installCatalogues {
		if !strings.HasPrefix(page, "https://") || !slices.Contains(knownManagers, manager) {
			t.Errorf("catalogue %s: bad page %q", manager, page)
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
