package scaffolding

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/monosmith-app/monosmith/libs/plugin-docker/images"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
)

// webMarkers are files a static web app's root holds: Vite's page,
// Angular's workspace, Astro's config.
var webMarkers = []string{"index.html", "angular.json", "astro.config.mjs"}

var goDirective = regexp.MustCompile(`(?m)^go\s+(\d+\.\d+)`)

type templateData struct {
	Dir, GoVersion, NodeVersion string
}

// nodeAppBuilds reports whether a directory is a Node app with a build
// script: its build is what the image installs, pruned.
func nodeAppBuilds(dir string) bool {
	raw, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return false
	}
	var manifest struct {
		Scripts map[string]string `json:"scripts"`
	}
	if json.Unmarshal(raw, &manifest) != nil {
		return false
	}
	_, ok := manifest.Scripts["build"]
	return ok
}

// Generate writes a new app's Dockerfile, after its language plugin wrote
// the app: a Go app builds its binary in Docker, a static web app is
// served by nginx, a Node app that builds installs from its pruned build.
// Another app gets a note: its Dockerfile is the author's to write. Libraries and packages do not ship as images.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	res := protocol.GenerateResult{Files: []string{}}
	if params.Kind != protocol.ProjectApp {
		return res, nil
	}
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Directory))
	if _, err := os.Stat(filepath.Join(dir, images.Dockerfile)); err == nil {
		return res, nil
	}
	data := templateData{Dir: params.Directory}
	var dockerfile string
	if raw, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
		data.GoVersion = "1.26"
		if m := goDirective.FindSubmatch(raw); m != nil {
			data.GoVersion = string(m[1])
		}
		dockerfile = goDockerfile
	}
	for _, marker := range webMarkers {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil && dockerfile == "" {
			dockerfile = webDockerfile
		}
	}
	if dockerfile == "" && nodeAppBuilds(dir) {
		data.NodeVersion = "24"
		dockerfile = nodeDockerfile
	}
	if dockerfile == "" {
		res.Notes = append(res.Notes, "Write "+params.Directory+"/"+images.Dockerfile+" (the context is the workspace) to ship "+params.Name+" as an image.")
		return res, nil
	}
	files := map[string]string{"[[.Dir]]/" + images.Dockerfile: dockerfile}
	if _, err := os.Stat(filepath.Join(workspaceRoot, ".dockerignore")); errors.Is(err, fs.ErrNotExist) {
		files[".dockerignore"] = dockerignore
	}
	written, err := templating.Write(workspaceRoot, files, data)
	if err != nil {
		return res, err
	}
	res.Files = written
	res.Notes = append(res.Notes, "smith run image -p "+params.Name+" builds its container image; a release pushes it to the [conventions.docker] registry (or registries) in smith.toml.")
	return res, nil
}
