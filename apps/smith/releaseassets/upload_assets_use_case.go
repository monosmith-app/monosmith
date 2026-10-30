package releaseassets

import (
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
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
		env := append(os.Environ(), "SMITH_PLATFORM="+platform, "SMITH_VERSION="+version, "SMITH_TAG="+tag)
		code, err := taskrun.RunProcess(ctx, taskrun.Spec{Argv: shellArgv(p.Release.Build), Dir: s.Root, Env: env, Out: req.Out})
		if err != nil || code != 0 {
			return done, problem.New(fmt.Sprintf("the build of %s for %s failed (exit %d): %v", p.Name, platform, code, err),
				"run it yourself from the workspace root: "+p.Release.Build)
		}
		files, err := assetFiles(os.DirFS(s.Root), s.Root, p.Release.Assets)
		if err != nil {
			return done, problem.New(fmt.Sprintf("%s for %s: %v", p.Name, platform, err),
				"make build leave the files, or fix assets in [projects."+p.Name+".release]")
		}
		repo, err := release.Repository(s)
		if err != nil {
			return done, err
		}
		if err := uploadFiles(ctx, s.Root, repo.String(), tag, files, req.Out); err != nil {
			return done, err
		}
		fmt.Fprintf(req.Out, "uploaded %d file(s) to %s\n", len(files), tag)
		done = append(done, p.Name)
	}
	return done, nil
}
