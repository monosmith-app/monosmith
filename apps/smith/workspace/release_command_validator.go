package workspace

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// validateReleaseCommand checks the [projects.<name>.release] of a project
// whose release files come from a command: it names what to build, which
// files to upload and known platforms, and leaves out what only applies to
// the archives a plugin builds.
func validateReleaseCommand(name string, po ProjectOverride, githubReleases bool) error {
	r := po.Release
	table := fmt.Sprintf("[projects.%s.release]", name)
	if strings.TrimSpace(r.Build) == "" {
		if r.Build != "" {
			return problem.New(table+" build is empty", "set build to the command that makes the files, e.g. build = \"make dist\"")
		}
		if len(r.Assets) > 0 {
			return problem.New(table+" lists assets but has no build, so nothing makes them",
				"add build = \"<the command that makes the files>\", or remove assets to let the project's plugin build its archives")
		}
		return nil
	}
	if len(r.Assets) == 0 {
		return problem.New(table+" has a build but no assets, so nothing would be uploaded",
			"add assets = [\"out/*.zip\"]: workspace-relative globs of the files the build leaves")
	}
	for _, a := range r.Assets {
		if strings.TrimSpace(a) == "" || !doublestar.ValidatePattern(a) {
			return problem.New(fmt.Sprintf("%s assets has %q, which is not a glob", table, a), "write workspace-relative globs such as \"out/*.zip\"")
		}
	}
	if len(r.Platforms) == 0 {
		return problem.New(table+" has a build but no platforms", "add platforms = [\"linux-x64\"]: the build runs once on a runner of each; "+strings.Join(CommandPlatforms, ", "))
	}
	for _, p := range r.Platforms {
		if !slices.Contains(CommandPlatforms, p) {
			return problem.New(fmt.Sprintf("%s platforms has %q, which is not a platform", table, p), "use one of "+strings.Join(CommandPlatforms, ", "))
		}
	}
	if !githubReleases {
		return problem.New(table+" uploads its files to the GitHub Release, which is not created",
			"set github_releases = true in [release]")
	}
	if slices.Contains(po.Tags, "native") {
		return problem.New(table+" has a build and the project is tagged native, which builds its assets on each OS another way",
			"remove the native tag, or remove build")
	}
	var plugin []string
	for _, s := range []struct {
		key string
		set bool
	}{
		{"asset_name", r.AssetName != ""}, {"os_alias", len(r.OSAlias) > 0}, {"arch_alias", len(r.ArchAlias) > 0},
		{"version_variable", r.VersionVariable != ""}, {"package_managers", len(r.PackageManagers) > 0},
		{"linux_packages", len(r.LinuxPackages) > 0}, {"npm", len(r.NPM) > 0}, {"pypi", r.PyPI != ""},
		{"nuget", r.NuGet != ""}, {"rubygems", r.RubyGems != ""}, {"pub", r.Pub != ""},
	} {
		if s.set {
			plugin = append(plugin, s.key)
		}
	}
	if len(plugin) > 0 {
		keys := strings.Join(plugin, ", ")
		return problem.New(fmt.Sprintf("%s sets %s next to build; those shape the archives smith builds itself, and a build command makes its own files", table, keys),
			"remove "+keys+", or remove build to let the plugin build the archives")
	}
	return nil
}
