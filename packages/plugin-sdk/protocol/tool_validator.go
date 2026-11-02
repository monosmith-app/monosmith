package protocol

import (
	"fmt"
	"slices"
	"strings"
)

// ToolManagers are the package managers a ToolInstall may name.
var ToolManagers = []string{"winget", "brew", "brew-cask", "apt", "dnf", "pacman", "npm", "uv", "fnm", "rustup"}

// ToolPinFormats are the formats a ToolVersionPin may name.
var ToolPinFormats = []string{"text", "go-mod", "pyproject", "global-json", "rust-toolchain"}

// ValidateToolNeed checks that a tool a plugin declares is complete: it can be
// detected, every install entry names a known manager, and it carries the
// official page, the source of the entries and the date they were checked.
//
// Parameters: t is the declaration.
// Returns: nil when it is well formed, else an error that names the problem.
func ValidateToolNeed(t ToolNeed) error {
	switch {
	case t.Name == "" || t.Title == "" || t.Why == "" || t.Consequence == "" || t.Program == "":
		return fmt.Errorf("tool %q: name, title, why, consequence and program are required", t.Name)
	case !strings.HasPrefix(t.Download, "https://") || !strings.HasPrefix(t.Source, "https://"):
		return fmt.Errorf("tool %s: download and source must be https addresses", t.Name)
	case len(t.Checked) != len("2006-01-02"):
		return fmt.Errorf("tool %s: checked must be a date (YYYY-MM-DD)", t.Name)
	}
	for _, in := range t.Install {
		if !slices.Contains(ToolManagers, in.Manager) || in.Package == "" {
			return fmt.Errorf("tool %s: an install entry needs a known manager (%s) and a package", t.Name, strings.Join(ToolManagers, ", "))
		}
	}
	for _, pin := range t.Pins {
		if pin.File == "" || !slices.Contains(ToolPinFormats, pin.Format) {
			return fmt.Errorf("tool %s: a pin needs a file and a format (%s)", t.Name, strings.Join(ToolPinFormats, ", "))
		}
	}
	return nil
}
