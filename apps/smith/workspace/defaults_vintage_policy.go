package workspace

import (
	"fmt"
	"regexp"
	"strings"
)

// DefaultsVintages lists, oldest first, every defaults vintage this smith
// knows. A vintage is the set of defaults smith applies to a workspace
// ([workspace] defaults, ADR 0054). A new vintage is added here, with its
// rules, when a default changes; the defaults of an older vintage never change.
var DefaultsVintages = []string{LatestDefaults}

// LatestDefaults is the newest vintage as a constant, for templates.
const LatestDefaults = "2026.11"

var vintagePattern = regexp.MustCompile(`^\d{4}\.(0[1-9]|1[0-2])$`)

// LatestDefaultsVintage is the newest vintage this smith knows: the one a new
// workspace records.
func LatestDefaultsVintage() string { return DefaultsVintages[len(DefaultsVintages)-1] }

// VintageState tells how a workspace's defaults vintage compares to this smith.
type VintageState int

const (
	// VintageCurrent means the workspace uses the latest vintage.
	VintageCurrent VintageState = iota
	// VintageBehind means the workspace uses an older vintage, or records none.
	VintageBehind
	// VintageAhead means the workspace records a vintage this smith does not know.
	VintageAhead
)

// DefaultsVintageState compares the vintage a workspace records with the
// newest this smith knows. An empty value is a workspace made before vintages
// existed: it is behind, whatever the latest is.
//
// Parameters: recorded is [workspace] defaults.
// Returns: the state of the workspace against the latest vintage.
func DefaultsVintageState(recorded string) VintageState {
	switch {
	case recorded == "":
		return VintageBehind
	case recorded == LatestDefaultsVintage():
		return VintageCurrent
	case recorded > LatestDefaultsVintage():
		return VintageAhead
	default:
		return VintageBehind
	}
}

// validateDefaultsVintage refuses a vintage that is not spelled YYYY.MM.
func validateDefaultsVintage(cfg *Config) error {
	d := strings.TrimSpace(cfg.Workspace.Defaults)
	if d == "" || vintagePattern.MatchString(d) {
		return nil
	}
	return fmt.Errorf("%s: [workspace] defaults %q is not a vintage: write the year and month, such as %q", FileName, cfg.Workspace.Defaults, LatestDefaultsVintage())
}
