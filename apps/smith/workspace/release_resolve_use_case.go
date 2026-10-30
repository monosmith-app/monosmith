package workspace

import (
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ResolveRelease completes and checks a project's [projects.<name>.release]
// table: a plugin's defaults (a Tauri app's build, assets and platforms) fill
// what the table leaves out, and the result is validated as a release from a
// command, or as the archives the project's plugin builds. The table is not
// changed; the resolved copy is returned.
func ResolveRelease(name string, po ProjectOverride, defaults *protocol.ReleaseDefaults, githubReleases bool) (*ProjectRelease, error) {
	if po.Release == nil {
		return nil, nil
	}
	r := *po.Release
	table := fmt.Sprintf("[projects.%s.release]", name)
	if defaults != nil && (r.Build == "" || strings.TrimSpace(r.Build) != "") && !slices.Contains(po.Tags, "native") && !setsArchiveOnly(&r) {
		if r.Build == "" {
			r.Build = defaults.Build
		}
		if len(r.Assets) == 0 {
			r.Assets = slices.Clone(defaults.Assets)
		}
		if len(r.Platforms) == 0 {
			r.Platforms = slices.Clone(defaults.Platforms)
		}
		if r.AssetName == "" {
			r.AssetName = defaults.AssetName
		}
		if r.Checksums == nil && defaults.Checksums {
			on := true
			r.Checksums = &on
		}
	}
	if err := validateReleaseCommand(name, &r, po.Tags, githubReleases); err != nil {
		return nil, err
	}
	if r.Build != "" {
		return &r, nil
	}
	if r.Checksums != nil {
		return nil, problem.New(table+" sets checksums, which only a release with a build command attaches",
			"remove checksums, or add build")
	}
	if len(r.Platforms) == 0 {
		r.Platforms = append([]string(nil), DefaultPlatforms...)
	}
	if r.AssetName == "" {
		r.AssetName = "{product}_{version}_{os}_{arch}.{ext}"
	}
	for _, p := range r.Platforms {
		if os, arch, ok := strings.Cut(p, "/"); !ok || os == "" || arch == "" {
			return nil, fmt.Errorf("%s: %s platform %q must be os/arch", FileName, table, p)
		}
	}
	return &r, nil
}

// setsArchiveOnly reports whether the table sets something only the archives
// a plugin builds have, which a plugin's build command defaults would replace.
func setsArchiveOnly(r *ProjectRelease) bool {
	return len(r.OSAlias) > 0 || len(r.ArchAlias) > 0 || r.VersionVariable != "" || len(r.PackageManagers) > 0 ||
		len(r.LinuxPackages) > 0 || len(r.NPM) > 0 || r.PyPI != "" || r.NuGet != "" || r.RubyGems != "" || r.Pub != ""
}
