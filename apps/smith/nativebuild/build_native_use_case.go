package nativebuild

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Tasks are what a native project runs on every runner.
var Tasks = []string{"lint", "test", "build", "package"}

// Run builds the native projects on this machine. On a Linux CI runner it
// first installs [ci] native_linux_packages. On a push to the release branch
// in GitHub Actions it then publishes this platform's build of every native
// package released at HEAD, and attaches this platform's build of every
// native app released at HEAD to the app's GitHub Release.
func Run(ctx context.Context, s *session.Session, out io.Writer) error {
	var native []*projectgraph.Project
	for _, p := range s.Graph.Projects {
		if release.Native(p) {
			native = append(native, p)
		}
	}
	if len(native) == 0 {
		fmt.Fprintln(out, "nothing to build natively: no project is tagged native")
		return nil
	}
	if packages := s.Config.CI.NativeLinuxPackages; len(packages) > 0 && runtime.GOOS == "linux" && inCI() {
		argv := append([]string{"sudo", "apt-get", "install", "-y", "--no-install-recommends"}, packages...)
		for _, cmd := range [][]string{{"sudo", "apt-get", "update"}, argv} {
			if code, err := taskrun.RunProcess(ctx, taskrun.Spec{Argv: cmd, Dir: s.Root, Env: os.Environ(), Out: out}); err != nil || code != 0 {
				return fmt.Errorf("%s failed (exit %d): %v", strings.Join(cmd, " "), code, err)
			}
		}
	}
	if err := os.MkdirAll(filepath.Join(s.Root, "dist", "drop"), 0o755); err != nil {
		return err
	}
	summary, err := taskrun.RunInWorkspace(ctx, s, taskrun.Request{Tasks: Tasks, Projects: []string{"tag:" + release.NativeTag}, Out: out})
	if err != nil {
		return err
	}
	fmt.Fprintln(out, summary.String())
	if !summary.OK() {
		return taskrun.ErrTasksFailed
	}
	github, azure := releasePush(s), azureReleasePush(s)
	if !github && !azure {
		return nil
	}
	if err := publishPackages(ctx, s, native, out); err != nil {
		return err
	}
	if !github {
		// Assets go to GitHub Releases, which an Azure pipeline does not create.
		return nil
	}
	return attach(ctx, s, native, out)
}

// publishPackages publishes this platform's build of every native package
// released at HEAD (a VS Code extension per target), and of every native app a
// deploy plugin ships (a VS Code extension app). The release job tagged
// it without publishing; the version is stamped into its manifest here, as
// the release job stamps it, and restored after.
func publishPackages(ctx context.Context, s *session.Session, native []*projectgraph.Project, out io.Writer) error {
	head, err := s.Git.RevParse("HEAD")
	if err != nil {
		return err
	}
	platform := runtime.GOOS + "/" + runtime.GOARCH
	for _, p := range native {
		if p.Kind != protocol.ProjectPackage && (p.Kind != protocol.ProjectApp || len(p.Deployers) == 0) {
			continue
		}
		tag, version := taggedAt(s, p, head)
		if tag == "" {
			fmt.Fprintf(out, "%s was not released by this commit; nothing to publish\n", p.Name)
			continue
		}
		if err := publishPlatform(ctx, s, p, version, platform, out); err != nil {
			return fmt.Errorf("publishing %s %s for %s: %w", p.Name, version, platform, err)
		}
	}
	return nil
}

func publishPlatform(ctx context.Context, s *session.Session, p *projectgraph.Project, version, platform string, out io.Writer) error {
	var vr protocol.VersionResult
	if _, err := s.Plugins.CallContext(ctx, p.Plugin, protocol.MethodVersion, protocol.VersionParams{Project: p.Project, Version: version}, &vr, time.Minute); err != nil {
		return fmt.Errorf("writing version: %w", err)
	}
	if len(vr.Files) > 0 {
		defer func() {
			if err := s.Git.RestoreFiles(vr.Files...); err != nil {
				fmt.Fprintf(out, "could not restore %v: %v\n", vr.Files, err)
			}
		}()
	}
	// A package is published by its own plugin; an app by the deploy plugins
	// that ship it.
	publishers := []string{p.Plugin}
	if p.Kind == protocol.ProjectApp {
		publishers = p.Deployers
	}
	for _, plugin := range publishers {
		var pr protocol.PublishResult
		params := protocol.PublishParams{Project: p.Project, Version: version, Platform: platform}
		if ok, err := s.Plugins.CallContext(ctx, plugin, protocol.MethodPublish, params, &pr, 30*time.Minute); err != nil {
			return err
		} else if !ok {
			pr.Skipped = "published by its tag"
		}
		switch {
		case pr.Published:
			fmt.Fprintf(out, "published %s %s for %s to %s\n", p.Name, version, platform, pr.Registry)
		case pr.Skipped != "":
			fmt.Fprintf(out, "%s %s for %s: %s\n", p.Name, version, platform, pr.Skipped)
		}
	}
	return nil
}

// attach adds this platform's assets to the GitHub Release of every native
// app tagged at HEAD.
func attach(ctx context.Context, s *session.Session, native []*projectgraph.Project, out io.Writer) error {
	head, err := s.Git.RevParse("HEAD")
	if err != nil {
		return err
	}
	for _, p := range native {
		if p.Release == nil {
			continue
		}
		tag, version := taggedAt(s, p, head)
		if tag == "" {
			fmt.Fprintf(out, "%s was not released by this commit; nothing to attach\n", p.Name)
			continue
		}
		spec := platformAssetSpec(p)
		if spec == nil {
			fmt.Fprintf(out, "%s builds its release files with its own command; the assets job attaches them\n", p.Name)
			continue
		}
		var built protocol.PublishResult
		params := protocol.PublishParams{Project: p.Project, Version: version, Assets: spec}
		if ok, err := s.Plugins.CallContext(ctx, p.Plugin, protocol.MethodPublish, params, &built, 30*time.Minute); err != nil || !ok {
			return fmt.Errorf("building %s's %s asset: %v", p.Name, spec.Platforms[0], err)
		}
		client, err := githubrelease.FromEnvironment()
		if err != nil {
			return err
		}
		repo, err := release.Repository(s)
		if err != nil {
			return err
		}
		files := make([]string, len(built.Assets))
		for i, a := range built.Assets {
			files[i] = filepath.Join(s.Root, filepath.FromSlash(a))
		}
		res, err := githubrelease.Ensure(client, repo, githubrelease.Request{Tag: tag, Title: p.Name + " " + version, Assets: files})
		if err != nil {
			return fmt.Errorf("GitHub Release %s: %w", tag, err)
		}
		fmt.Fprintf(out, "attached %s to %s\n", strings.Join(res.Uploaded, ", "), tag)
	}
	return nil
}

// platformAssetSpec is what to build for p on this platform, or nil when p
// makes its release files with its own build command ([projects.x.release]
// build): the plugin builds no archives for it, and the assets job runs the
// command and attaches the files.
func platformAssetSpec(p *projectgraph.Project) *protocol.AssetSpec {
	spec := release.AssetSpec(p)
	if spec == nil {
		return nil
	}
	spec.Platforms = []string{runtime.GOOS + "/" + runtime.GOARCH}
	return spec
}

// taggedAt is the project's release tag pointing at commit, and its version.
func taggedAt(s *session.Session, p *projectgraph.Project, commit string) (string, string) {
	return release.TaggedAt(s, p, commit)
}

func inCI() bool { return os.Getenv("GITHUB_ACTIONS") == "true" || os.Getenv("TF_BUILD") != "" }

// releasePush reports a GitHub Actions push to the release branch.
func releasePush(s *session.Session) bool {
	return os.Getenv("GITHUB_ACTIONS") == "true" && os.Getenv("GITHUB_EVENT_NAME") == "push" && os.Getenv("GITHUB_REF_NAME") == s.Config.Release.Branch
}

// azureReleasePush reports an Azure Pipelines run for a push to the release
// branch: the same condition the generated pipeline's release step has.
func azureReleasePush(s *session.Session) bool {
	reason := os.Getenv("BUILD_REASON")
	return os.Getenv("TF_BUILD") != "" && (reason == "IndividualCI" || reason == "BatchedCI") && os.Getenv("BUILD_SOURCEBRANCH") == "refs/heads/"+s.Config.Release.Branch
}
