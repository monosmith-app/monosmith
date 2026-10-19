package turboplan

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/turboworkspace"
)

// Specifier is a dependency range after conversion.
type Specifier struct {
	// Range is what npm accepts, "" when none could be found.
	Range string
	// Problem says why no range could be found.
	Problem string
	// Approximate is set when the range is looser than the original: the
	// package's version was not known, so it is "*".
	Approximate bool
}

// SpecifierOf translates a dependency range npm does not accept: pnpm's and
// yarn's `workspace:` protocol, and pnpm's `catalog:`. Any other range is
// returned as it is, with ok false.
func SpecifierOf(w *turboworkspace.Workspace, dependency, spec string) (Specifier, bool) {
	switch {
	case strings.HasPrefix(spec, "workspace:"):
		rest := strings.TrimPrefix(spec, "workspace:")
		switch rest {
		case "*":
			return Specifier{Range: "*"}, true
		case "^", "~":
			if target := w.PackageByName(dependency); target != nil && target.JSON.Version != "" {
				return Specifier{Range: rest + target.JSON.Version}, true
			}
			return Specifier{Range: "*", Approximate: true}, true
		}
		if strings.HasPrefix(rest, ".") || strings.HasPrefix(rest, "/") || strings.Contains(rest, "@") && !strings.HasPrefix(rest, "^") && !strings.HasPrefix(rest, "~") {
			return Specifier{Problem: spec + " names a path or an aliased package; npm finds a workspace package by its name and version range"}, true
		}
		return Specifier{Range: rest}, true
	case strings.HasPrefix(spec, "catalog:"):
		name := strings.TrimPrefix(spec, "catalog:")
		if w.Pnpm != nil {
			catalog := w.Pnpm.Catalog
			if name != "" && name != "default" {
				catalog = w.Pnpm.Catalogs[name]
			}
			if v, ok := catalog[dependency]; ok {
				return Specifier{Range: v}, true
			}
		}
		return Specifier{Problem: spec + " has no entry for " + dependency + " in pnpm-workspace.yaml (a yarn catalog in .yarnrc.yml is not read)"}, true
	case strings.HasPrefix(spec, "link:") || strings.HasPrefix(spec, "portal:") || strings.HasPrefix(spec, "patch:"):
		return Specifier{Problem: spec + " is a " + strings.SplitN(spec, ":", 2)[0] + ": specifier, which npm does not accept; use file: for a local directory"}, true
	}
	return Specifier{}, false
}
