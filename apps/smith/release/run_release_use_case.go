package release

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/linuxpackages"
	"github.com/monosmith-app/monosmith/apps/smith/npmbinaries"
	"github.com/monosmith-app/monosmith/apps/smith/nugettool"
	"github.com/monosmith-app/monosmith/apps/smith/packagemanifests"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/pypiwheels"
	"github.com/monosmith-app/monosmith/apps/smith/rubygems"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/splitrepo"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Options configure a release run.
type Options struct {
	DryRun bool
	// Push sends tags to the release remote (when it exists).
	Push bool
	Out  io.Writer
}

// Execute releases a plan. Each package is stamped, built, published, restored
// and tagged in dependency order; it stops at the first failure, keeping what
// already finished.
func Execute(ctx context.Context, s *session.Session, plan Plan, opts Options) (Report, error) {
	report := Report{DryRun: opts.DryRun, Started: time.Now().UTC(), Unchanged: plan.Unchanged, Results: []ReportResult{}}
	defer func() { report.Finished = time.Now().UTC(); writeReport(s.Root, report) }()

	if len(plan.Items) == 0 {
		return report, nil
	}
	if !opts.DryRun {
		if err := Preflight(s, plan); err != nil {
			report.Error = err.Error()
			return report, err
		}
		pushPending(s, opts)
	}
	versions := map[string]string{}
	for _, item := range plan.Items {
		for _, e := range s.Graph.Get(item.Project).Exports {
			versions[e] = item.Next
		}
	}
	for _, item := range plan.Items {
		res := ReportResult{Item: item, Status: "planned"}
		if err := releaseOne(ctx, s, item, versions, opts, &res); err != nil {
			res.Status, res.Error = "failed", err.Error()
			report.Results = append(report.Results, res)
			report.Error = item.Project + ": " + err.Error()
			return report, errors.New(report.Error)
		}
		report.Results = append(report.Results, res)
	}
	return report, nil
}

func releaseOne(ctx context.Context, s *session.Session, item Item, versions map[string]string, opts Options, res *ReportResult) error {
	p := s.Graph.Get(item.Project)
	say := func(format string, args ...any) { fmt.Fprintf(opts.Out, item.Project+": "+format+"\n", args...) }
	if commit := s.Git.TagCommit(item.Tag); commit != "" {
		res.Status = "already-released"
		say("%s already exists", item.Tag)
		if opts.DryRun || !s.Config.Release.GitHubReleases {
			return nil
		}
		if err := pushTag(s, item, opts, res); err != nil {
			return err
		}
		if err := mirror(s, p, item, opts, res, say); err != nil {
			return err
		}
		if err := publishGitHubRelease(s, p, item, res, say, nil); err != nil {
			return err
		}
		return dispatchPub(s, p, item, res, say)
	}

	deps := map[string]string{}
	for _, d := range s.Graph.TransitiveDeps(p.Name) {
		for _, e := range s.Graph.Get(d).Exports {
			if v, ok := versions[e]; ok {
				deps[e] = v
			}
		}
	}
	var vr protocol.VersionResult
	ok, err := s.Plugins.Call(p.Plugin, protocol.MethodVersion, protocol.VersionParams{Project: p.Project, Version: item.Next, Dependencies: deps, DryRun: opts.DryRun}, &vr, time.Minute)
	if err != nil {
		return fmt.Errorf("writing version: %w", err)
	}
	if ok {
		res.Files = vr.Files
	}
	if !opts.DryRun && len(vr.Files) > 0 {
		defer func() {
			if err := s.Git.RestoreFiles(vr.Files...); err != nil {
				say("could not restore %v: %v", vr.Files, err)
			}
		}()
	}

	// An app's assets are built by publishing; its own build is not needed.
	if !opts.DryRun && p.Release == nil {
		if _, has := p.Tasks["build"]; has {
			say("building %s", item.Next)
			summary, err := taskrun.RunInWorkspace(ctx, s, taskrun.Request{Tasks: []string{"build"}, Projects: []string{p.Name}, Out: opts.Out, NoCache: len(vr.Files) > 0})
			if err != nil {
				return err
			}
			if !summary.OK() {
				return errors.New("build failed")
			}
		}
	}

	if p.Release != nil && !opts.DryRun {
		say("building assets for %s", item.Next)
	}
	pr, err := publish(s, p, item, opts.DryRun)
	if err != nil {
		return err
	}
	res.Published, res.Registry, res.Publish, res.Assets = pr.Published, pr.Registry, pr.Skipped, pr.Assets
	if err := deploy(s, p, item, opts.DryRun, say); err != nil {
		return err
	}
	if err := distribute(s, p, item, pr, opts.DryRun, res, say); err != nil {
		return err
	}
	packages, err := linuxPackages(s, p, item, pr, opts.DryRun)
	if err != nil {
		return err
	}
	if opts.DryRun {
		say("would release %s (%s)", item.Next, item.Reason)
		if attach := append(slices.Clone(pr.Assets), packages...); len(attach) > 0 {
			say("would attach %s", strings.Join(attach, ", "))
		}
		if p.Release != nil && p.Release.Pub != "" && !Native(p) {
			say("would publish %s %s to pub.dev from the tag %s", p.Release.Pub, item.Next, PubTag(p, item.Next))
		}
		return mirror(s, p, item, opts, res, say)
	}

	if err := s.Git.CreateTag(item.Tag, fmt.Sprintf("%s %s", item.Project, item.Next)); err != nil {
		return err
	}
	res.Status = "released"
	say("tagged %s", item.Tag)
	if err := pushTag(s, item, opts, res); err != nil {
		return err
	}
	if err := mirror(s, p, item, opts, res, say); err != nil {
		return err
	}
	if s.Config.Release.GitHubReleases {
		assets := append(append(pr.Assets, packages...), packageManifests(s, p, item, pr, say)...)
		if err := publishGitHubRelease(s, p, item, res, say, assets); err != nil {
			return err
		}
		return dispatchPub(s, p, item, res, say)
	}
	return nil
}

// packageManifests writes the package manager manifests a release asks for
// beside its archives, to attach to the GitHub Release; a problem is said,
// never fatal: the release itself is done.
func packageManifests(s *session.Session, p *projectgraph.Project, item Item, pr protocol.PublishResult, say func(string, ...any)) []string {
	if p.Release == nil || len(p.Release.PackageManagers) == 0 || Native(p) || len(pr.AssetPlatforms) == 0 {
		return nil
	}
	repo, err := Repository(s)
	if err != nil {
		say("no package manager manifests: %v", err)
		return nil
	}
	product := p.Release.Product
	if product == "" {
		product = p.Name
	}
	binary := pr.Binary
	if binary == "" {
		binary = p.Name
	}
	r := packagemanifests.Release{Product: product, Binary: binary, Version: item.Next, Description: p.Release.Description, Homepage: p.Release.Homepage, License: p.Release.License, Owner: repo.Owner, Repo: repo.Name, Tag: item.Tag}
	written, skipped, err := packagemanifests.Write(s.Root, packagemanifests.Built{Assets: pr.Assets, Platforms: pr.AssetPlatforms}, r, p.Release.PackageManagers)
	for _, why := range skipped {
		say("no %s manifest", why)
	}
	if err != nil {
		say("package manager manifests: %v", err)
	}
	return written
}

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

// NativeTag marks a project built on each OS (smith ci native), whose release
// assets each native job attaches for its own platform.
const NativeTag = "native"

// Native reports whether a project is built on each OS.
func Native(p *projectgraph.Project) bool { return slices.Contains(p.Tags, NativeTag) }

// publish asks the project's plugin to publish it; an app with a release
// configuration gets its assets built instead, except a native app's, which
// the native jobs build and attach. A native package is published by the
// native jobs too, each its own platform's build (smith ci native).
func publish(s *session.Session, p *projectgraph.Project, item Item, dryRun bool) (protocol.PublishResult, error) {
	if Native(p) && p.Kind == protocol.ProjectPackage {
		return protocol.PublishResult{Skipped: "published for each platform by the native jobs"}, nil
	}
	spec := AssetSpec(p)
	if Native(p) {
		spec = nil
	}
	params := protocol.PublishParams{Project: p.Project, Version: item.Next, DryRun: dryRun, Assets: spec}
	var pr protocol.PublishResult
	ok, err := s.Plugins.Call(p.Plugin, protocol.MethodPublish, params, &pr, 30*time.Minute)
	if err != nil {
		return pr, fmt.Errorf("publishing: %w", err)
	}
	if !ok {
		if params.Assets != nil {
			return pr, fmt.Errorf("the %s plugin cannot build release assets", p.Plugin)
		}
		pr.Skipped = "published by its tag"
	}
	return pr, nil
}

// PlannedAssets lists an app's release assets for a version, with each
// archive's platform and the binary inside, without building them.
func PlannedAssets(s *session.Session, p *projectgraph.Project, version string) (protocol.PublishResult, error) {
	return publish(s, p, Item{Project: p.Name, Next: version}, true)
}

// deploy publishes the project with each deploy plugin that ships it (a
// container image to push), after its own plugin published it.
func deploy(s *session.Session, p *projectgraph.Project, item Item, dryRun bool, say func(string, ...any)) error {
	for _, plugin := range p.Deployers {
		params := protocol.PublishParams{Project: p.Project, Version: item.Next, DryRun: dryRun}
		var pr protocol.PublishResult
		ok, err := s.Plugins.Call(plugin, protocol.MethodPublish, params, &pr, 30*time.Minute)
		switch {
		case err != nil:
			return fmt.Errorf("deploying with %s: %w", plugin, err)
		case !ok:
		case pr.Published:
			say("%s published %s to %s", plugin, item.Next, pr.Registry)
		case pr.Skipped != "":
			say("%s: %s", plugin, pr.Skipped)
		}
	}
	return nil
}

// distribute publishes an app's binaries as npm packages, PyPI wheels, a
// .NET tool and gems, made from the archives its plugin built, before the
// app is tagged: a failure stops the release, and the re-run skips what the
// registries have.
func distribute(s *session.Session, p *projectgraph.Project, item Item, pr protocol.PublishResult, dryRun bool, res *ReportResult, say func(string, ...any)) error {
	r := p.Release
	if r == nil || (len(r.NPM) == 0 && r.PyPI == "" && r.NuGet == "" && r.RubyGems == "") {
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
	if len(r.NPM) > 0 {
		release := npmbinaries.Release{Names: r.NPM, Binary: binary, Version: item.Next, Description: r.Description, License: r.License, Homepage: homepage, Repository: repository, OutDir: out + "/npm"}
		for _, a := range archives {
			release.Archives = append(release.Archives, npmbinaries.Archive{Path: a.path, OS: a.os, Arch: a.arch})
		}
		result, err := npmbinaries.Publish(s.Root, release, dryRun)
		report("npm", result.Published, result.Existing, result.Skipped)
		if err != nil {
			return fmt.Errorf("npm: %w", err)
		}
	}
	if r.PyPI != "" {
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
	if r.NuGet != "" {
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
	if r.RubyGems != "" {
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

// AssetSpec is what a plugin needs to build an app's release assets, or nil
// for a project without them.
func AssetSpec(p *projectgraph.Project) *protocol.AssetSpec {
	r := p.Release
	if r == nil {
		return nil
	}
	return &protocol.AssetSpec{
		Name:            r.AssetName,
		Product:         r.Product,
		Platforms:       r.Platforms,
		OSAlias:         r.OSAlias,
		ArchAlias:       r.ArchAlias,
		VersionVariable: r.VersionVariable,
		OutDir:          "dist/release/" + p.Name,
	}
}

// mirror pushes a project with a split repository there, once its tag is
// pushed: the mirror's history is taken from the monorepo's tag, so a tag
// that only exists locally is not mirrored yet, and the re-run that pushes
// it mirrors it too.
func mirror(s *session.Session, p *projectgraph.Project, item Item, opts Options, res *ReportResult, say func(string, ...any)) error {
	if p.SplitRepo == "" {
		return nil
	}
	if !opts.DryRun && !res.Pushed {
		say("not mirroring to %s: the tag was not pushed", p.SplitRepo)
		return nil
	}
	out, err := splitrepo.Mirror(s.Git, splitrepo.Request{Prefix: p.Root, Ref: item.Tag, Repo: p.SplitRepo, Version: item.Next, Token: splitrepo.Token(), DryRun: opts.DryRun})
	if err != nil {
		return err
	}
	res.Mirror = out.Mirror
	switch {
	case opts.DryRun:
		say("would mirror %s to %s", p.Root, out.Mirror)
	case out.Skipped != "":
		say("%s", out.Skipped)
	default:
		say("mirrored %s to %s", p.Root, out.Mirror)
	}
	return nil
}

func pushTag(s *session.Session, item Item, opts Options, res *ReportResult) error {
	if !opts.Push || !s.Git.HasRemote(s.Config.Release.Remote) {
		return nil
	}
	if err := s.Git.PushTag(s.Config.Release.Remote, item.Tag); err != nil {
		return fmt.Errorf("pushing %s: %w (the tag exists locally; re-run to push it)", item.Tag, err)
	}
	res.Pushed = true
	return nil
}

// publishGitHubRelease creates the tag's GitHub Release and attaches the
// assets. When assets is nil (a resumed release), the plugin lists the
// expected files and they are rebuilt only if the release lacks any.
func publishGitHubRelease(s *session.Session, p *projectgraph.Project, item Item, res *ReportResult, say func(string, ...any), assets []string) error {
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
	if assets == nil && (p.Release != nil || attachesItsFile(p)) && !Native(p) {
		existing, found, err := client.ReleaseByTag(repo, item.Tag)
		if err != nil {
			return err
		}
		planned, err := publish(s, p, item, true)
		if err != nil {
			return err
		}
		packages, err := linuxPackages(s, p, item, planned, true)
		if err != nil {
			return err
		}
		if !found || missing(existing, append(planned.Assets, packages...)) {
			say("building assets for %s", item.Next)
			built, err := publish(s, p, item, false)
			if err != nil {
				return err
			}
			packages, err := linuxPackages(s, p, item, built, false)
			if err != nil {
				return err
			}
			assets = append(built.Assets, packages...)
		}
	}
	files := make([]string, len(assets))
	for i, a := range assets {
		files[i] = filepath.Join(s.Root, filepath.FromSlash(a))
	}
	out, err := githubrelease.Ensure(client, repo, githubrelease.Request{Tag: item.Tag, Title: item.Project + " " + item.Next, Notes: RenderNotes(item), Assets: files})
	if err != nil {
		return fmt.Errorf("GitHub Release %s: %w (re-run the release to finish it)", item.Tag, err)
	}
	res.GitHubRelease = out.URL
	if out.Created {
		say("created GitHub Release %s", out.URL)
	}
	if len(out.Uploaded) > 0 {
		say("attached %s", strings.Join(out.Uploaded, ", "))
	}
	return nil
}

// attachesItsFile reports whether a project's release has an asset without
// the project having a [release] table: a recipe attaches its own file, so a
// URL can pin a version of it. A resumed release lists it from the plugin.
func attachesItsFile(p *projectgraph.Project) bool { return p.Language == "recipe" }

func missing(r githubrelease.Release, assets []string) bool {
	have := map[string]bool{}
	for _, a := range r.Assets {
		have[a.Name] = true
	}
	for _, a := range assets {
		if !have[path.Base(a)] {
			return true
		}
	}
	return false
}

// Repository is the GitHub repository releases go to: GITHUB_REPOSITORY on
// Actions, otherwise the release remote's URL.
func Repository(s *session.Session) (githubrelease.Repository, error) {
	if env := os.Getenv("GITHUB_REPOSITORY"); env != "" {
		return githubrelease.ParseRepository(env)
	}
	remote, err := s.Git.RemoteURL(s.Config.Release.Remote)
	if err != nil {
		return githubrelease.Repository{}, fmt.Errorf("no GitHub repository: set GITHUB_REPOSITORY or add the %s remote", s.Config.Release.Remote)
	}
	return githubrelease.ParseRepository(remote)
}

// pushPending pushes release tags at HEAD left local by an interrupted run.
func pushPending(s *session.Session, opts Options) {
	if !opts.Push || !s.Git.HasRemote(s.Config.Release.Remote) {
		return
	}
	head, err := s.Git.RevParse("HEAD")
	if err != nil {
		return
	}
	out, err := s.Git.Run("tag", "--points-at", head)
	if err != nil || out == "" {
		return
	}
	for _, tag := range strings.Fields(out) {
		_ = s.Git.PushTag(s.Config.Release.Remote, tag)
	}
}

func writeReport(root string, r Report) {
	dir := filepath.Join(root, ".smith")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	raw, err := json.MarshalIndent(r, "", "  ")
	if err == nil {
		_ = os.WriteFile(filepath.Join(dir, "release-report.json"), append(raw, '\n'), 0o644)
	}
	if summary := os.Getenv("GITHUB_STEP_SUMMARY"); summary != "" {
		if f, err := os.OpenFile(summary, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644); err == nil {
			_, _ = f.WriteString(r.Markdown() + "\n")
			_ = f.Close()
		}
	}
}
