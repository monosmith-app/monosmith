package releasing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/libs/plugin-dotnet/registries"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// DotnetRunner runs dotnet; tests replace it.
var DotnetRunner = func(dir string, argv []string) error { return toolrun.Run(dir, argv) }

// projectFile is the project's .csproj/.fsproj, project-relative.
func projectFile(workspaceRoot string, p protocol.Project) (string, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
	for _, pattern := range []string{"*.csproj", "*.fsproj"} {
		if found, _ := filepath.Glob(filepath.Join(dir, pattern)); len(found) > 0 {
			return filepath.Base(found[0]), nil
		}
	}
	return "", fmt.Errorf("%s has no project file", p.Root)
}

// WriteVersion records a release's version in the project file.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	file, err := projectFile(workspaceRoot, params.Project)
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

// NuGetSource is where packages are pushed; NUGET_SOURCE replaces it.
const NuGetSource = "https://api.nuget.org/v3/index.json"

// Publish packs a package into the workspace's dist/<project dir> and pushes
// it with dotnet nuget push (NUGET_API_KEY); a version the feed already has
// is skipped, so a re-run is safe. Libraries release by their tag; an app with
// release assets gets a zip per platform instead.
func Publish(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	if params.Assets != nil {
		return BuildAssets(workspaceRoot, params)
	}
	if params.Project.Kind != protocol.ProjectPackage {
		return protocol.PublishResult{Skipped: "released by its tag"}, nil
	}
	source := os.Getenv("NUGET_SOURCE")
	if source == "" {
		source = NuGetSource
	}
	key := os.Getenv("NUGET_API_KEY")
	// A feed smith.toml names for the package wins, with its own key.
	if feed, feedKey, ok := registries.Push(params.Project, os.Getenv); ok {
		source, key = feed, feedKey
	}
	result := protocol.PublishResult{Registry: source}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	file, err := projectFile(workspaceRoot, params.Project)
	if err != nil {
		return result, err
	}
	out := filepath.Join(workspaceRoot, filepath.FromSlash(protocol.ProjectBuildDir(params.Project.Root)))
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root))
	if err := DotnetRunner(dir, []string{"dotnet", "pack", file, "--nologo", "-c", "Release", "-o", out, "-p:Version=" + params.Version}); err != nil {
		return result, err
	}
	id := params.Project.Name
	if len(params.Project.Exports) > 0 {
		id = params.Project.Exports[0]
	}
	nupkg := filepath.Join(out, id+"."+params.Version+".nupkg")
	argv := []string{"dotnet", "nuget", "push", nupkg, "--source", source, "--skip-duplicate"}
	if key != "" {
		argv = append(argv, "--api-key", key)
	}
	if err := DotnetRunner(workspaceRoot, argv); err != nil {
		return result, err
	}
	result.Published = true
	return result, nil
}

// NuGetIndex is nuget.org's package index (flat container); tests replace it.
var NuGetIndex = "https://api.nuget.org/v3-flatcontainer"

// Published asks nuget.org for a package's newest stable version.
func Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	if params.Project.Kind != protocol.ProjectPackage || len(params.Project.Exports) == 0 {
		return protocol.PublishedResult{}, nil
	}
	if r := params.Project.Registries; r != nil && r.Publish != nil && r.Publish.NuGet != "" {
		return protocol.PublishedResult{Registry: r.Publish.NuGet, Unreachable: true, Detail: "smith does not ask a private NuGet feed for its versions yet"}, nil
	}
	result := protocol.PublishedResult{Registry: "nuget.org"}
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Get(NuGetIndex + "/" + strings.ToLower(params.Project.Exports[0]) + "/index.json")
	if err != nil {
		result.Unreachable, result.Detail = true, err.Error()
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
	var index struct{ Versions []string }
	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		result.Unreachable, result.Detail = true, err.Error()
		return result, nil
	}
	for _, v := range index.Versions {
		if !strings.Contains(v, "-") && (result.Version == "" || newer(v, result.Version)) {
			result.Version = v
		}
	}
	return result, nil
}

func newer(a, b string) bool {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		var x, y int
		if i < len(as) {
			x, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			y, _ = strconv.Atoi(bs[i])
		}
		if x != y {
			return x > y
		}
	}
	return false
}
