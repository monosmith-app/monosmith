package release

import (
	"fmt"
	"maps"
	"path"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/linuxpackages"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// linuxPackages builds the .deb and .rpm packages a release asks for from
// its linux archives, before the app is tagged: a failure stops the release.
// A dry run returns the files it would build.
func linuxPackages(s *session.Session, p *projectgraph.Project, item Item, pr protocol.PublishResult, dryRun bool) ([]string, error) {
	r, ok := linuxRelease(s, p, item, pr)
	if !ok {
		return nil, nil
	}
	built := linuxpackages.Built{Assets: pr.Assets, Platforms: pr.AssetPlatforms}
	if dryRun {
		dir := ""
		for asset := range pr.AssetPlatforms {
			dir = path.Dir(asset)
		}
		var names []string
		for _, name := range linuxpackages.Expected(r, slices.Collect(maps.Values(pr.AssetPlatforms))) {
			names = append(names, path.Join(dir, name))
		}
		return names, nil
	}
	packages, err := linuxpackages.Build(s.Root, built, r)
	if err != nil {
		return nil, fmt.Errorf("linux packages: %w", err)
	}
	return packages, nil
}

// linuxRelease is what a project's Linux packages hold, if it asks for any.
func linuxRelease(s *session.Session, p *projectgraph.Project, item Item, pr protocol.PublishResult) (linuxpackages.Release, bool) {
	if p.Release == nil || len(p.Release.LinuxPackages) == 0 || Native(p) || len(pr.AssetPlatforms) == 0 {
		return linuxpackages.Release{}, false
	}
	product := p.Release.Product
	if product == "" {
		product = p.Name
	}
	binary := pr.Binary
	if binary == "" {
		binary = p.Name
	}
	homepage, maintainer := p.Release.Homepage, product
	if repo, err := Repository(s); err == nil {
		if homepage == "" {
			homepage = "https://github.com/" + repo.String()
		}
		maintainer = repo.Owner + " <https://github.com/" + repo.String() + ">"
	}
	return linuxpackages.Release{Kinds: p.Release.LinuxPackages, Product: product, Binary: binary, Version: item.Next, Summary: p.Release.Description, Homepage: homepage, License: p.Release.License, Maintainer: maintainer}, true
}
