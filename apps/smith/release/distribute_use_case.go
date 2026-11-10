package release

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/npmbinaries"
	"github.com/monosmith-app/monosmith/apps/smith/nugettool"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/pypiwheels"
	"github.com/monosmith-app/monosmith/apps/smith/rubygems"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/npmpropagation"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// distribute publishes an app's binaries as npm packages, PyPI wheels, a
// .NET tool and gems, made from the archives its plugin built, before the
// app is tagged: a failure stops the release, and the re-run skips what the
// registries have.
func distribute(s *session.Session, p *projectgraph.Project, item Item, pr protocol.PublishResult, dryRun bool, res *ReportResult, say func(string, ...any)) error {
	r := p.Release
	if r == nil || (len(r.NPM) == 0 && r.PyPI == "" && r.NuGet == "" && r.RubyGems == "") {
		return nil
	}
	if PublishesToNoRegistry(p) {
		say("no npm, PyPI, NuGet or RubyGems packages for %s: smith.toml sends it to no registry", p.Name)
		return nil
	}
	if Native(p) {
		say("no npm, PyPI, NuGet or RubyGems packages: a native app's archives are built on separate runners")
		return nil
	}
	binary := pr.Binary
	if binary == "" {
		binary = p.Name
	}
	homepage, repository := r.Homepage, ""
	if repo, err := Repository(s); err == nil {
		repository = "https://github.com/" + repo.String()
		if homepage == "" {
			homepage = repository
		}
	}
	type archive struct{ path, os, arch string }
	var archives []archive
	for asset, platform := range pr.AssetPlatforms {
		goos, goarch, _ := strings.Cut(platform, "/")
		archives = append(archives, archive{asset, goos, goarch})
	}
	out := "dist/release/" + p.Name
	verb := "published"
	if dryRun {
		verb = "would publish"
	}
	report := func(kind string, published, existing, skipped []string) {
		res.Packages = append(res.Packages, published...)
		res.ExistingPackages = append(res.ExistingPackages, existing...)
		if len(published) > 0 {
			say("%s %s %s", verb, kind, strings.Join(published, ", "))
		}
		if len(existing) > 0 {
			say("%s already has %s", kind, strings.Join(existing, ", "))
		}
		if len(skipped) > 0 {
			say("no %s package for %s", kind, strings.Join(skipped, ", "))
		}
	}
	// skip says, and records, that stable_only leaves a pre-release out of a
	// registry.
	skip := func(kind, key, name string) bool {
		if !r.SkipsPrerelease(key, item.Next) {
			return false
		}
		res.SkippedPackages = append(res.SkippedPackages, kind+" "+name+": skipped: pre-release")
		say("%s %s: skipped: pre-release %s (stable_only in smith.toml)", kind, name, item.Next)
		return true
	}
	if len(r.NPM) > 0 {
		release := npmbinaries.Release{Names: r.NPM, Binary: binary, Version: item.Next, Description: r.Description, License: r.License, Homepage: homepage, Repository: repository, OutDir: out + "/npm"}
		if ch := ChannelOf(s.Config.Release, s.Git.CurrentBranch()); ch != nil {
			release.Tag = ch.Tag
		}
		for _, a := range archives {
			release.Archives = append(release.Archives, npmbinaries.Archive{Path: a.path, OS: a.os, Arch: a.arch})
		}
		wait, err := npmpropagation.ParseWait(PropagationWait(s.Config, p))
		if err != nil {
			return err
		}
		release.PropagationWait = wait
		release.Progress = func(line string) { say("%s", line) }
		result, err := npmbinaries.Publish(s.Root, release, dryRun)
		report("npm", result.Published, result.Existing, result.Skipped)
		for _, move := range result.PendingMoves {
			say("the dist-tag %s of %s moves at the end of the release, once npm serves the platform packages", move.Tag, move.Package)
		}
		res.addPending("", result.PendingMoves...)
		if err != nil {
			return fmt.Errorf("npm: %w", err)
		}
	}
	if r.PyPI != "" && !skip("PyPI", "pypi", r.PyPI) {
		release := pypiwheels.Release{Name: r.PyPI, Binary: binary, Version: item.Next, Description: r.Description, License: r.License, Homepage: homepage, OutDir: out + "/pypi"}
		for _, a := range archives {
			release.Archives = append(release.Archives, pypiwheels.Archive{Path: a.path, OS: a.os, Arch: a.arch})
		}
		result, err := pypiwheels.Publish(s.Root, release, dryRun)
		report("PyPI", result.Published, result.Existing, result.Skipped)
		if err != nil {
			return fmt.Errorf("PyPI: %w", err)
		}
	}
	if r.NuGet != "" && !skip("NuGet", "nuget", r.NuGet) {
		release := nugettool.Release{ID: r.NuGet, Binary: binary, Version: item.Next, Description: r.Description, License: r.License, Homepage: homepage, Repository: repository, OutDir: out + "/nuget"}
		if repo, err := Repository(s); err == nil {
			release.Authors = repo.Owner
		}
		for _, a := range archives {
			release.Archives = append(release.Archives, nugettool.Archive{Path: a.path, OS: a.os, Arch: a.arch})
		}
		result, err := nugettool.Publish(s.Root, release, dryRun)
		report("NuGet", result.Published, result.Existing, result.Skipped)
		if err != nil {
			return fmt.Errorf("NuGet: %w", err)
		}
	}
	if r.RubyGems != "" && !skip("RubyGems", "rubygems", r.RubyGems) {
		release := rubygems.Release{Name: r.RubyGems, Binary: binary, Version: item.Next, Description: r.Description, License: r.License, Homepage: homepage, Repository: repository, OutDir: out + "/rubygems"}
		for _, a := range archives {
			release.Archives = append(release.Archives, rubygems.Archive{Path: a.path, OS: a.os, Arch: a.arch})
		}
		result, err := rubygems.Publish(s.Root, release, dryRun)
		report("RubyGems", result.Published, result.Existing, result.Skipped)
		if err != nil {
			return fmt.Errorf("RubyGems: %w", err)
		}
	}
	return nil
}
