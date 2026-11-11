package release

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/packagemanifests"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
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
	// Terse leaves out the per-package "would release" line of a dry run,
	// which the plan above it already says.
	Terse bool
	// Registries asks the registries before the first publish whether they
	// would accept the release; nil reads the environment.
	Registries *RegistryProbe
}

// Execute releases a plan. Each package is stamped, built, published, restored
// and tagged in dependency order; it stops at the first failure, keeping what
// already finished.
func Execute(ctx context.Context, s *session.Session, plan Plan, opts Options) (Report, error) {
	started := time.Now().UTC()
	report, err := execute(ctx, s, plan, opts)
	report.Started, report.Finished = started, time.Now().UTC()
	for i := range report.Results {
		report.Results[i].Targets = TargetsOf(report.Results[i])
	}
	writeReport(s.Root, report)
	recordDeployments(s, report, opts.Out)
	return report, err
}

func execute(ctx context.Context, s *session.Session, plan Plan, opts Options) (Report, error) {
	report := Report{DryRun: opts.DryRun, Channel: plan.Channel, Unchanged: plan.Unchanged, NeverReleased: plan.NeverReleased, Results: []ReportResult{}}

	if len(plan.Items) == 0 {
		return report, nil
	}
	if !opts.DryRun {
		if err := Preflight(s, plan); err != nil {
			report.Error = err.Error()
			return report, err
		}
		if err := checkRegistries(s, plan, opts.Registries, opts.Out); err != nil {
			report.Error = err.Error()
			return report, err
		}
		pushPending(s, opts)
	}
	versions, err := SiblingVersions(s.Git, s.Graph, plan)
	if err != nil {
		report.Error = err.Error()
		return report, err
	}
	var failures []string
	for _, item := range plan.Items {
		res := ReportResult{Item: item, Status: "planned"}
		began := time.Now()
		err := releaseOne(ctx, s, item, versions, opts, &res)
		res.TookMs = time.Since(began).Milliseconds()
		if err != nil {
			res.Status, res.Error = "failed", err.Error()
			report.Results = append(report.Results, res)
			journalRelease(s, res, opts.Out)
			failures = append(failures, item.Project+": "+err.Error())
			if opts.DryRun {
				// A dry run changes nothing, so every project's plan is shown
				// and the run still fails at the end.
				continue
			}
			report.Error = failures[0]
			return report, errors.New(report.Error)
		}
		report.Results = append(report.Results, res)
		journalRelease(s, res, opts.Out)
	}
	if len(failures) > 0 {
		report.Error = strings.Join(failures, "; ")
		return report, errors.New(report.Error)
	}
	if err := attachBundle(s, &report, opts); err != nil {
		report.Error = err.Error()
		return report, err
	}
	// The dist-tags move last, once every project has published: the wait for
	// npm blocks no publish, and no short-lived credential of a later project
	// (a federated sign-in, a trusted-publishing token) runs out meanwhile.
	if !opts.DryRun {
		say := func(format string, args ...any) { fmt.Fprintf(opts.Out, "smith: "+format+"\n", args...) }
		if err := FinalizeTags(ctx, s, &report, say); err != nil {
			report.Error = err.Error()
			return report, err
		}
	}
	if len(report.PendingTags()) > 0 {
		return report, ErrPropagationPending
	}
	return report, nil
}

func releaseOne(ctx context.Context, s *session.Session, item Item, versions map[string]string, opts Options, res *ReportResult) error {
	p := s.Graph.Get(item.Project)
	say := func(format string, args ...any) { fmt.Fprintf(opts.Out, item.Project+": "+format+"\n", args...) }
	if commit := s.Git.TagCommit(item.Tag); commit != "" {
		res.Status = "already-released"
		say("%s already exists", item.Tag)
		if !opts.DryRun {
			if err := resumeNPM(ctx, s, p, item, versions, opts, res, say); err != nil {
				return err
			}
		}
		if opts.DryRun || !s.Config.Release.GitHubReleases {
			return nil
		}
		if err := pushTag(s, item, opts, res); err != nil {
			return err
		}
		if err := mirror(s, p, item, opts, res, say); err != nil {
			return err
		}
		if PacksToRelease(p) {
			return resumePack(ctx, s, p, item, versions, opts, res, say)
		}
		if err := publishGitHubRelease(ctx, s, p, item, res, say, nil); err != nil {
			return err
		}
		return dispatchPub(s, p, item, res, say)
	}

	stamped, restore, err := stampVersion(ctx, s, p, item, versions, opts, res, say)
	if err != nil {
		return err
	}
	defer restore()

	// An app's assets are built by publishing; its own build is not needed.
	if !opts.DryRun && p.Release == nil {
		if err := buildProject(ctx, s, p, item, opts, len(stamped) > 0, say); err != nil {
			return err
		}
	}

	if p.Release != nil && !opts.DryRun {
		say("building assets for %s", item.Next)
	}
	pr, err := publish(ctx, s, p, item, opts.DryRun)
	if err != nil {
		return err
	}
	res.Published, res.Registry, res.Publish, res.Assets = pr.Published, pr.Registry, pr.Skipped, pr.Assets
	if len(p.Exports) > 0 {
		res.Package = p.Exports[0]
	}
	for _, note := range pr.Notes {
		say("note: %s", note)
		res.Publish = strings.TrimSpace(res.Publish + " " + note)
	}
	res.addPending(p.Plugin, pr.PendingMoves...)
	if PacksToRelease(p) && len(pr.Assets) > 0 {
		if err := sealPackage(s, &pr, opts.DryRun, res); err != nil {
			return err
		}
	}
	if BuildsAssetsFromCommand(p) {
		say("%s are made by `%s` on a runner each (smith ci assets, or smith release assets on this machine) and uploaded to the GitHub Release: %s",
			strings.Join(p.Release.Platforms, ", "), p.Release.Build, strings.Join(p.Release.Assets, ", "))
	}
	if err := deploy(ctx, s, p, item, opts.DryRun, res, say); err != nil {
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
		if !opts.Terse {
			say("would release %s (%s)", item.Next, item.Reason)
		}
		if attach := append(slices.Clone(pr.Assets), packages...); len(attach) > 0 {
			say("would attach %s", strings.Join(attach, ", "))
		}
		if p.Release != nil && p.Release.Pub != "" && !Native(p) {
			say("would publish %s %s to pub.dev from the tag %s", p.Release.Pub, item.Next, PubTag(p, item.Next))
			if head, err := s.Git.RevParse("HEAD"); err == nil {
				if elsewhere, ok := pubTagElsewhere(s, PubTag(p, item.Next), head); ok {
					say("warning: %s already exists on %s, not on this commit: the release would stop at pub.dev", PubTag(p, item.Next), short(elsewhere))
				}
			}
		}
		return mirror(s, p, item, opts, res, say)
	}

	if err := s.Git.CreateTag(item.Tag, TagMessage(item)); err != nil {
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
		if err := publishGitHubRelease(ctx, s, p, item, res, say, assets); err != nil {
			return err
		}
		return dispatchPub(s, p, item, res, say)
	}
	return nil
}

// buildProject runs the project's build task before it is published or packed,
// in its release profile (smith build --release, ADR 0092). noCache is set when the version was stamped into its files, which the cache
// key does not see.
func buildProject(ctx context.Context, s *session.Session, p *projectgraph.Project, item Item, opts Options, noCache bool, say func(string, ...any)) error {
	if _, has := p.Tasks["build"]; !has {
		return nil
	}
	say("building %s", item.Next)
	summary, err := taskrun.RunInWorkspace(ctx, s, taskrun.Request{Tasks: []string{"build"}, Projects: []string{p.Name}, Out: opts.Out, NoCache: noCache, Release: true})
	if err != nil {
		return err
	}
	if !summary.OK() {
		return errors.New("build failed")
	}
	return nil
}

// stampVersion writes the release version into the project's files and the
// convention files. The returned function puts the files back as they were; it
// does nothing in a dry run or when nothing was written.
func stampVersion(ctx context.Context, s *session.Session, p *projectgraph.Project, item Item, versions map[string]string, opts Options, res *ReportResult, say func(string, ...any)) ([]string, func(), error) {
	deps := map[string]string{}
	for _, d := range s.Graph.TransitiveDeps(p.Name) {
		for _, e := range s.Graph.Get(d).Exports {
			if v, ok := versions[e]; ok {
				deps[e] = v
			}
		}
	}
	var vr protocol.VersionResult
	ok, err := s.Plugins.CallContext(ctx, p.Plugin, protocol.MethodVersion, protocol.VersionParams{Project: p.Project, Version: item.Next, Dependencies: deps, DryRun: opts.DryRun}, &vr, time.Minute)
	if err != nil {
		return nil, func() {}, fmt.Errorf("writing version: %w", err)
	}
	extra, err := WriteConventionVersions(ctx, s, p, item.Next, opts.DryRun)
	if err != nil {
		return nil, func() {}, fmt.Errorf("writing version: %w", err)
	}
	vr.Files = append(vr.Files, extra...)
	if ok || len(extra) > 0 {
		res.Files = vr.Files
	}
	restore := func() {}
	if !opts.DryRun && len(vr.Files) > 0 {
		restore = func() {
			if err := s.Git.RestoreFiles(vr.Files...); err != nil {
				say("could not restore %v: %v", vr.Files, err)
			}
		}
	}
	return vr.Files, restore, nil
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

// publish asks the project's plugin to publish it; an app with a release
// configuration gets its assets built instead, except a native app's, which
// the native jobs build and attach. A native package is published by the
// native jobs too, each its own platform's build (smith ci native).
func publish(ctx context.Context, s *session.Session, p *projectgraph.Project, item Item, dryRun bool) (protocol.PublishResult, error) {
	if Native(p) && p.Kind == protocol.ProjectPackage {
		return protocol.PublishResult{Skipped: "published for each platform by the native jobs"}, nil
	}
	spec := AssetSpec(p)
	planOnly := false
	if Native(p) {
		// A native app's archives are built by the native jobs, one platform
		// each. A dry run still asks for this machine's, to say what a native
		// job would attach and to check that it can be built.
		planOnly = dryRun && spec != nil
		if planOnly {
			spec.Platforms = []string{runtime.GOOS + "/" + runtime.GOARCH}
		} else {
			spec = nil
		}
	}
	if spec == nil && p.Kind == protocol.ProjectApp {
		// The same for every language: an app is released by its tag, its release
		// files and its deploy plugins. To publish it to a registry, make it a package.
		return protocol.PublishResult{Skipped: "an app is released by its tag, its files and its deploy plugins, never to a registry"}, nil
	}
	if spec == nil && p.Registries != nil && p.Registries.NoPublish {
		return protocol.PublishResult{Skipped: "published by its tag (registry = \"none\")"}, nil
	}
	if spec == nil && p.PublishTo == "none" {
		return protocol.PublishResult{Skipped: "published by its tag (the manifest says it is never published)"}, nil
	}
	if spec == nil && PacksToRelease(p) {
		return packToRelease(ctx, s, p, item, dryRun)
	}
	params := protocol.PublishParams{Project: p.Project, Version: item.Next, DryRun: dryRun, Assets: spec, PropagationWait: PropagationWait(s.Config, p)}
	if ch := ChannelOf(s.Config.Release, s.Git.CurrentBranch()); ch != nil {
		params.Tag = ch.Tag
	}
	var pr protocol.PublishResult
	ok, err := s.Plugins.CallContext(ctx, p.Plugin, protocol.MethodPublish, params, &pr, 30*time.Minute)
	if err != nil {
		return pr, fmt.Errorf("publishing: %w", err)
	}
	if !ok {
		if params.Assets != nil && !planOnly {
			return pr, fmt.Errorf("the %s plugin cannot build release assets", p.Plugin)
		}
		pr.Skipped = "published by its tag"
	}
	return pr, nil
}

// PlannedAssets lists an app's release assets for a version, with each
// archive's platform and the binary inside, without building them.
func PlannedAssets(s *session.Session, p *projectgraph.Project, version string) (protocol.PublishResult, error) {
	return publish(context.Background(), s, p, Item{Project: p.Name, Next: version}, true)
}

// deploy publishes the project with each deploy plugin that ships it (a
// container image to push), after its own plugin published it.
func deploy(ctx context.Context, s *session.Session, p *projectgraph.Project, item Item, dryRun bool, res *ReportResult, say func(string, ...any)) error {
	commit, _ := s.Git.RevParse("HEAD")
	branch := s.Config.Release.Branch
	for _, plugin := range p.Deployers {
		params := protocol.PublishParams{Project: p.Project, Version: item.Next, DryRun: dryRun, Commit: commit, Branch: branch}
		var pr protocol.PublishResult
		ok, err := s.Plugins.CallContext(ctx, plugin, protocol.MethodPublish, params, &pr, 30*time.Minute)
		switch {
		case err != nil:
			return fmt.Errorf("deploying with %s: %w", plugin, err)
		case !ok:
		case pr.Published:
			say("%s published %s to %s", plugin, item.Next, pr.Registry)
			for _, image := range pr.Images {
				res.Images = append(res.Images, image)
				say("  %s", ImageLine(image))
			}
		case pr.Skipped != "":
			say("%s: %s", plugin, pr.Skipped)
		}
	}
	return nil
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
