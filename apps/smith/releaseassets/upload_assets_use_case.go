package releaseassets

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Request says what to build and upload.
type Request struct {
	// Platform is the one this machine is a runner of. Default: this machine's.
	Platform string
	// Project limits the run to one project. Default: every project that has
	// a build command and was released at HEAD.
	Project string
	Out     io.Writer
}

// Run builds, for the platform, every project released at HEAD whose release
// has a build command, and uploads its files to the project's GitHub Release.
// It returns the projects it uploaded for.
func Run(ctx context.Context, s *session.Session, req Request) ([]string, error) {
	platform := req.Platform
	if platform == "" {
		platform = workspace.HostCommandPlatform()
	}
	if !slices.Contains(workspace.CommandPlatforms, platform) {
		return nil, problem.New(fmt.Sprintf("%q is not a platform smith builds release files for", platform),
			"smith release assets --platform "+strings.Join(workspace.CommandPlatforms, " | "))
	}
	head, err := s.Git.RevParse("HEAD")
	if err != nil {
		return nil, err
	}
	var done []string
	for _, p := range s.Graph.Projects {
		if !release.BuildsAssetsFromCommand(p) || (req.Project != "" && p.Name != req.Project) {
			continue
		}
		if !slices.Contains(p.Release.Platforms, platform) {
			fmt.Fprintf(req.Out, "%s has no %s build: its platforms are %s\n", p.Name, platform, strings.Join(p.Release.Platforms, ", "))
			continue
		}
		tag, version := release.TaggedAt(s, p, head)
		if tag == "" {
			fmt.Fprintf(req.Out, "%s was not released by this commit; nothing to build\n", p.Name)
			continue
		}
		fmt.Fprintf(req.Out, "building %s %s for %s: %s\n", p.Name, version, platform, p.Release.Build)
		n, err := buildAndUpload(ctx, s, req, p, platform, tag, version)
		if err != nil {
			return done, err
		}
		fmt.Fprintf(req.Out, "uploaded %d file(s) to %s\n", n, tag)
		done = append(done, p.Name)
	}
	return done, nil
}

// buildAndUpload stamps the release's version into the project's version files
// (a Tauri app's tauri.conf.json and Cargo.toml, so the installers carry it),
// runs the build, names and uploads the files, and puts the version files back.
// It returns how many files it uploaded.
func buildAndUpload(ctx context.Context, s *session.Session, req Request, p *projectgraph.Project, platform, tag, version string) (int, error) {
	if stamped := stampVersions(ctx, s, p, version); len(stamped) > 0 {
		defer func() { _ = s.Git.RestoreFiles(stamped...) }()
	}
	env := append(os.Environ(), "SMITH_PLATFORM="+platform, "SMITH_VERSION="+version, "SMITH_TAG="+tag)
	code, err := taskrun.RunProcess(ctx, taskrun.Spec{Argv: shellArgv(p.Release.Build), Dir: s.Root, Env: env, Out: req.Out})
	if err != nil || code != 0 {
		return 0, problem.New(fmt.Sprintf("the build of %s for %s failed (exit %d): %v", p.Name, platform, code, err),
			"run it yourself from the workspace root: "+p.Release.Build)
	}
	files, err := assetFiles(os.DirFS(s.Root), s.Root, p.Release.Assets)
	if err != nil {
		return 0, problem.New(fmt.Sprintf("%s for %s: %v", p.Name, platform, err),
			"make build leave the files, or fix assets in [projects."+p.Name+".release]")
	}
	r := p.Release
	product := r.Product
	if product == "" {
		product = p.Name
	}
	if checksums := r.Checksums != nil && *r.Checksums; r.AssetName != "" || checksums {
		names, err := assetNames(files, r.AssetName, product, version, platform)
		if err != nil {
			return 0, problem.New(fmt.Sprintf("%s for %s: %v", p.Name, platform, err), "fix asset_name or assets in [projects."+p.Name+".release]")
		}
		sums := ""
		if checksums {
			sums = fmt.Sprintf("%s-%s-%s.sha256", product, version, platform)
		}
		staging := filepath.Join(s.Root, protocol.BuildDir, "release", p.Name+"-assets", platform)
		if files, err = stageAssets(staging, files, names, sums); err != nil {
			return 0, err
		}
	}
	repo, err := release.Repository(s)
	if err != nil {
		return 0, err
	}
	if err := uploadFiles(ctx, s.Root, repo.String(), tag, files, req.Out); err != nil {
		return 0, err
	}
	return len(files), nil
}

// stampVersions writes version into the project's version files with its
// plugins and returns the files it changed. A plugin that cannot is skipped:
// the build still runs with the versions the checkout holds.
func stampVersions(ctx context.Context, s *session.Session, p *projectgraph.Project, version string) []string {
	var files []string
	var vr protocol.VersionResult
	if ok, err := s.Plugins.CallContext(ctx, p.Plugin, protocol.MethodVersion, protocol.VersionParams{Project: p.Project, Version: version}, &vr, time.Minute); err == nil && ok {
		files = append(files, vr.Files...)
	}
	extra, _ := release.WriteConventionVersions(ctx, s, p, version, false)
	return append(files, extra...)
}
