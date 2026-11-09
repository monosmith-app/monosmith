package releasing

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-dart/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// PubRunner runs dart and flutter; tests replace it.
var PubRunner = func(dir string, argv []string) error { return toolrun.Run(dir, argv) }

func pubspec(workspaceRoot string, p protocol.Project) (string, discovery.Pubspec, error) {
	rel := path.Join(p.Root, "pubspec.yaml")
	src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(rel)))
	if err != nil {
		return rel, discovery.Pubspec{}, err
	}
	parsed, err := discovery.ParsePubspec(src)
	return rel, parsed, err
}

// WriteVersion records a release's version in the pubspec.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	rel := path.Join(params.Project.Root, "pubspec.yaml")
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

// PubDev is pub.dev's API; tests replace it.
var PubDev = "https://pub.dev"

// registry is where a package publishes: its publish_to, else pub.dev.
func registry(p discovery.Pubspec) string {
	if strings.HasPrefix(p.PublishTo, "http") {
		return strings.TrimSuffix(p.PublishTo, "/")
	}
	return PubDev
}

// Publish publishes a package with pub. pub.dev takes GitHub Actions' OIDC
// token (setup-dart sets it up); a repository the pubspec's publish_to names
// takes PUB_TOKEN. Apps and libraries release by their tag.
func Publish(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	if params.Assets != nil {
		return protocol.PublishResult{Skipped: "Dart release assets are not built yet: the app is released by its tag"}, nil
	}
	if params.Project.Kind != protocol.ProjectPackage {
		return protocol.PublishResult{Skipped: "released by its tag"}, nil
	}
	_, pub, err := pubspec(workspaceRoot, params.Project)
	if err != nil {
		return protocol.PublishResult{}, err
	}
	if pub.PublishTo == "none" {
		return protocol.PublishResult{Skipped: "publish_to: none, so the package is released by its tag"}, nil
	}
	result := protocol.PublishResult{Registry: registry(pub)}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	if result.Registry == PubDev && !params.Tagged {
		// pub.dev takes a package only from a GitHub Actions run on a tag
		// that matches the pattern set on the package, with GitHub's OIDC
		// token; the release job is neither, so the release starts that run.
		result.Skipped = "published to pub.dev by the pub workflow, on the package's pub tag"
		return result, nil
	}
	tool := "dart"
	if pub.Flutter {
		tool = "flutter"
	}
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root))
	if result.Registry != PubDev && os.Getenv("PUB_TOKEN") != "" {
		if err := PubRunner(dir, []string{"dart", "pub", "token", "add", result.Registry, "--env-var", "PUB_TOKEN"}); err != nil {
			return result, err
		}
	}
	if err := PubRunner(dir, []string{tool, "pub", "publish", "--force"}); err != nil {
		return result, err
	}
	result.Published = true
	return result, nil
}

// Published asks the package's repository (pub.dev by default) for its
// newest version.
func Published(workspaceRoot string, params protocol.PublishedParams) (protocol.PublishedResult, error) {
	if params.Project.Kind != protocol.ProjectPackage || len(params.Project.Exports) == 0 {
		return protocol.PublishedResult{}, nil
	}
	_, pub, err := pubspec(workspaceRoot, params.Project)
	if err != nil {
		return protocol.PublishedResult{}, err
	}
	if pub.PublishTo == "none" {
		return protocol.PublishedResult{}, nil
	}
	result := protocol.PublishedResult{Registry: registry(pub)}
	resp, err := (&http.Client{Timeout: protocol.RegistryLookupTimeout}).Get(result.Registry + "/api/packages/" + params.Project.Exports[0])
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
	var body struct {
		Latest struct{ Version string } `json:"latest"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		result.Unreachable, result.Detail = true, protocol.LookupFailure(err)
		return result, nil
	}
	result.Version = body.Latest.Version
	return result, nil
}
