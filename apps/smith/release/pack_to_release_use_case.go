package release

import (
	"context"
	"fmt"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/releasepack"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// attachedToRelease is what the report says a packed package went to.
const attachedToRelease = "attached to GitHub Release"

// packToRelease asks the project's plugin to pack the project (registry =
// "github-release"): it writes the package file and publishes nowhere. A
// language with no package file (Go, PHP, JVM, Dart) is released by its tag
// and keeps its tag-based behaviour. A plugin that cannot pack is an error and
// is never asked to publish, since it would publish to the public registry.
func packToRelease(ctx context.Context, s *session.Session, p *projectgraph.Project, item Item, dryRun bool) (protocol.PublishResult, error) {
	if _, ok := releasepack.EcosystemOf(p.Plugin); !ok {
		return protocol.PublishResult{Skipped: "released by its tag: the " + p.Plugin + " plugin has no package file to attach"}, nil
	}
	if c := s.Plugins.Get(p.Plugin); c == nil || !c.Info.Capabilities.Packs {
		return protocol.PublishResult{}, fmt.Errorf("the %s plugin cannot pack its package for a GitHub Release; update the plugin", p.Plugin)
	}
	var pr protocol.PublishResult
	params := protocol.PublishParams{Project: p.Project, Version: item.Next, DryRun: dryRun, PackDir: PackDir(p.Name)}
	if _, err := s.Plugins.CallContext(ctx, p.Plugin, protocol.MethodPublish, params, &pr, 30*time.Minute); err != nil {
		return pr, fmt.Errorf("packing: %w", err)
	}
	pr.Published = false
	return pr, nil
}

// sealPackage checksums the files a pack wrote and adds the checksum files to
// the assets, after all the package files: a release that holds a package file
// without its checksum file is an attach that stopped. A dry run only says
// what would be attached.
func sealPackage(s *session.Session, pr *protocol.PublishResult, dryRun bool, res *ReportResult) error {
	if dryRun {
		res.Publish = "would attach to GitHub Release"
		return nil
	}
	files, sidecars, err := releasepack.Checksum(s.Root, pr.Assets)
	if err != nil {
		return fmt.Errorf("checksum: %w", err)
	}
	res.Packed = files
	pr.Assets = append(pr.Assets, sidecars...)
	res.Assets = pr.Assets
	res.Publish = attachedToRelease
	return nil
}

// resumePack finishes the package files of a version whose tag exists. A
// release that holds every file and checksum is read back, not packed again;
// otherwise the project is stamped, built and packed from the checkout (which
// must be at the tag) and the files the release lacks are attached. "Already
// has the version" means the release for the tag already has the asset.
func resumePack(ctx context.Context, s *session.Session, p *projectgraph.Project, item Item, versions map[string]string, opts Options, res *ReportResult, say func(string, ...any)) error {
	eco, packs := releasepack.EcosystemOf(p.Plugin)
	if !packs {
		return publishGitHubRelease(ctx, s, p, item, res, say, nil)
	}
	if !res.Pushed {
		say("not creating a GitHub Release: the tag was not pushed")
		return nil
	}
	client, err := githubrelease.FromEnvironment()
	if err != nil {
		return err
	}
	repo, err := Repository(s)
	if err != nil {
		return err
	}
	existing, found, err := client.ReleaseByTag(repo, item.Tag)
	if err != nil {
		return err
	}
	if found && releasepack.Complete(existing, eco) {
		files, err := releasepack.Recover(client, repo, existing, eco)
		if err != nil {
			return fmt.Errorf("GitHub Release %s: %w", item.Tag, err)
		}
		res.Packed, res.GitHubRelease, res.Publish = files, existing.HTMLURL, attachedToRelease
		say("%s already has its package files", item.Tag)
		return nil
	}
	if head, _ := s.Git.RevParse("HEAD"); s.Git.TagCommit(item.Tag) != head {
		return fmt.Errorf("the GitHub Release %s lacks package files and HEAD is newer than its tag: check out the tag and run the release again", item.Tag)
	}
	say("packing %s again: its GitHub Release lacks package files", item.Tag)
	stamped, restore, err := stampVersion(ctx, s, p, item, versions, opts, res, say)
	if err != nil {
		return err
	}
	defer restore()
	if err := buildProject(ctx, s, p, item, opts, len(stamped) > 0, say); err != nil {
		return err
	}
	pr, err := publish(ctx, s, p, item, false)
	if err != nil {
		return err
	}
	if err := sealPackage(s, &pr, false, res); err != nil {
		return err
	}
	return publishGitHubRelease(ctx, s, p, item, res, say, pr.Assets)
}

// attachBundle attaches the run's bundle (release-bundle.json and INSTALL.md)
// to the GitHub Release of every package the run packed, once all of them are
// attached, so each release lists the whole set. A dry run says it would.
func attachBundle(s *session.Session, report *Report, opts Options) error {
	say := func(format string, args ...any) { fmt.Fprintf(opts.Out, "smith: "+format+"\n", args...) }
	var packages []releasepack.Package
	planned := 0
	for _, res := range report.Results {
		if p := s.Graph.Get(res.Project); p != nil && PacksToRelease(p) {
			planned++
		}
		eco, ok := releasepack.EcosystemOf(res.Plugin)
		if !ok || len(res.Packed) == 0 {
			continue
		}
		packages = append(packages, releasepack.Package{Project: res.Project, Ecosystem: eco.Name, Version: res.Next, Tag: res.Tag, Files: res.Packed})
	}
	if opts.DryRun {
		if planned > 0 {
			say("would attach release-bundle.json and INSTALL.md to the GitHub Release of each package, listing the whole set")
		}
		return nil
	}
	if len(packages) == 0 {
		return nil
	}
	client, err := githubrelease.FromEnvironment()
	if err != nil {
		return err
	}
	repo, err := Repository(s)
	if err != nil {
		return err
	}
	tags, err := releasepack.AttachBundle(client, repo, s.Root, releasepack.NewBundle(repo.String(), packages))
	if err != nil {
		return fmt.Errorf("release bundle: %w (re-run the release to finish it)", err)
	}
	say("attached release-bundle.json and INSTALL.md to %d GitHub Releases", len(tags))
	return nil
}
