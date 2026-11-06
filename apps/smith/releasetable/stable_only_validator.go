package releasetable

import (
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// validateStableOnly checks stable_only of [projects.<name>.release]: each
// entry names a registry that takes it, once, and the table publishes there.
func validateStableOnly(table string, r *workspace.ProjectRelease) error {
	published := map[string]bool{"pypi": r.PyPI != "", "nuget": r.NuGet != "", "rubygems": r.RubyGems != ""}
	seen := map[string]bool{}
	for _, name := range r.StableOnly {
		switch {
		case !slices.Contains(workspace.StableOnlyRegistries, name):
			return problem.New(fmt.Sprintf("%s stable_only has %q, which cannot skip pre-releases", table, name),
				"use "+strings.Join(workspace.StableOnlyRegistries, ", ")+" (npm keeps pre-releases under the channel's tag)")
		case seen[name]:
			return problem.New(fmt.Sprintf("%s stable_only lists %q twice", table, name), "list each registry once")
		case !published[name]:
			return problem.New(fmt.Sprintf("%s stable_only names %s, but the table sets no %s package", table, name, name),
				fmt.Sprintf("add %s = \"<package name>\", or remove %s from stable_only", name, name))
		}
		seen[name] = true
	}
	return nil
}
