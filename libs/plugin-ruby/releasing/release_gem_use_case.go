package releasing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// GemRunner runs gem; tests replace it.
var GemRunner = func(dir string, argv []string) error { return toolrun.Run(dir, argv) }

// gemspec is the project's gemspec, project-relative.
func gemspec(workspaceRoot string, p protocol.Project) (string, error) {
	found, _ := filepath.Glob(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "*.gemspec"))
	if len(found) == 0 {
		return "", fmt.Errorf("%s has no gemspec", p.Root)
	}
	return filepath.Base(found[0]), nil
}

// WriteVersion records a release's version in the gemspec.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	file, err := gemspec(workspaceRoot, params.Project)
	if err != nil {
		return protocol.VersionResult{}, err
	}
	rel := path.Join(params.Project.Root, file)
	full := filepath.Join(workspaceRoot, filepath.FromSlash(rel))
	src, err := os.ReadFile(full)
	if err != nil {
		return protocol.VersionResult{}, err
	}
	out, changed := SetVersion(string(src), params.Version)
	if !changed {
		return protocol.VersionResult{Files: []string{}}, nil
	}
	if !params.DryRun {
		if err := os.WriteFile(full, []byte(out), 0o644); err != nil {
			return protocol.VersionResult{}, err
		}
	}
	return protocol.VersionResult{Files: []string{rel}}, nil
}

// RubyGems is the gem server; GEM_HOST replaces it, tests too.
var RubyGems = "https://rubygems.org"

func host() string {
	if h := os.Getenv("GEM_HOST"); h != "" {
		return h
	}
	return RubyGems
}

// Publish packs a package into the workspace's dist/<project dir> and pushes
// it with gem push, which reads GEM_HOST_API_KEY. Apps and libraries release
// by their tag.
func Publish(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	if params.Assets != nil {
		return protocol.PublishResult{Skipped: "Ruby release assets are not built yet: the app is released by its tag"}, nil
	}
	if params.Project.Kind != protocol.ProjectPackage {
		return protocol.PublishResult{Skipped: "released by its tag"}, nil
	}
	result := protocol.PublishResult{Registry: host()}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	file, err := gemspec(workspaceRoot, params.Project)
	if err != nil {
		return result, err
	}
	out := filepath.Join(workspaceRoot, filepath.FromSlash(protocol.ProjectBuildDir(params.Project.Root)))
	if err := os.MkdirAll(out, 0o755); err != nil {
		return result, err
	}
	name := params.Project.Name
	if len(params.Project.Exports) > 0 {
		name = params.Project.Exports[0]
	}
	packed := filepath.Join(out, name+"-"+GemVersion(params.Version)+".gem")
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root))
	if err := GemRunner(dir, []string{"gem", "build", file, "--output", packed}); err != nil {
		return result, err
	}
	if err := GemRunner(dir, []string{"gem", "push", packed, "--host", result.Registry}); err != nil {
		return result, err
	}
	result.Published = true
	return result, nil
}

// Published asks the gem server for a gem's newest version.
func Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	if params.Project.Kind != protocol.ProjectPackage || len(params.Project.Exports) == 0 {
		return protocol.PublishedResult{}, nil
	}
	result := protocol.PublishedResult{Registry: host()}
	resp, err := (&http.Client{Timeout: protocol.RegistryLookupTimeout}).Get(result.Registry + "/api/v1/versions/" + params.Project.Exports[0] + "/latest.json")
	if err != nil {
		result.Unreachable, result.Detail = true, protocol.LookupFailure(err)
		return result, nil
	}
	defer func() { _ = resp.Body.Close() }()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return result, nil
	case resp.StatusCode != http.StatusOK:
		result.Unreachable, result.Detail = true, resp.Status
		return result, nil
	}
	var body struct{ Version string }
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		result.Unreachable, result.Detail = true, protocol.LookupFailure(err)
		return result, nil
	}
	// RubyGems answers "unknown" for a gem it does not have.
	if body.Version != "unknown" {
		result.Version = body.Version
	}
	return result, nil
}
