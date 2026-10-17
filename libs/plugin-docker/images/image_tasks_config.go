package images

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TaskName is the task that builds an app's image. It is not build, so
// lint, test and build never need a Docker daemon.
const TaskName = "image"

// Dockerfile is the file that makes an app ship as an image.
const Dockerfile = "Dockerfile"

// PruneTask is the task a language plugin gives an app whose image installs
// from a standalone deploy output (the node plugin's prune: a production
// package.json, lockfile and workspace_modules/ in the build directory). The
// image is built after it.
const PruneTask = "prune"

// Name is the image's name: the project's, lowercase as registries require.
func Name(p protocol.Project) string { return strings.ToLower(p.Name) }

// Ships reports whether a project ships as an image: an app with a
// Dockerfile in its root.
func Ships(workspaceRoot string, p protocol.Project) bool {
	if p.Kind != protocol.ProjectApp {
		return false
	}
	_, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), Dockerfile))
	return err == nil
}

// BuildArgs are docker build's arguments for the project's Dockerfile, with
// the workspace as the context (a Dockerfile copies the app's build from
// dist/, or builds from the sources it needs), tagged with each tag.
func BuildArgs(p protocol.Project, tags ...string) []string {
	argv := []string{"docker", "build", "-f", path.Join(p.Root, Dockerfile)}
	for _, t := range tags {
		argv = append(argv, "-t", t)
	}
	return append(argv, ".")
}

// Contribute gives every app that ships an image task, run from the
// workspace root after the app's build. It is never cached: the image
// lives in the Docker daemon, not in files smith can restore.
func Contribute(workspaceRoot string, params protocol.ContributeParams) protocol.ContributeResult {
	res := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, p := range params.Projects {
		if !Ships(workspaceRoot, p) {
			continue
		}
		version := p.Version
		if version == "" {
			version = "dev"
		}
		task := protocol.Task{
			Command:     BuildArgs(p, Name(p)+":"+version, Name(p)+":latest"),
			Cwd:         ".",
			Description: "Build the app's container image (" + Name(p) + ":" + version + ")",
		}
		for _, before := range []string{"build", PruneTask} {
			if _, ok := p.Tasks[before]; ok {
				task.DependsOn = append(task.DependsOn, before)
			}
		}
		res.Projects[p.Name] = protocol.Contribution{Tasks: map[string]protocol.Task{TaskName: task}}
	}
	return res
}
