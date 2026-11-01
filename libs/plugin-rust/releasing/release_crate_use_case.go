package releasing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// CargoRunner runs cargo; tests replace it.
var CargoRunner = func(dir string, argv []string) error { return toolrun.Run(dir, argv) }

// WriteVersion records a release's version in the crate's Cargo.toml and the
// requirement on it in every dependency the plugin was told of.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	rel := params.Project.Root + "/Cargo.toml"
	path := filepath.Join(workspaceRoot, filepath.FromSlash(rel))
	src, err := os.ReadFile(path)
	if err != nil {
		return protocol.VersionResult{}, err
	}
	out, changed := SetPackageVersion(string(src), params.Version)
	out, depsChanged := SetDependencyVersions(out, params.Dependencies)
	if !changed && !depsChanged {
		return protocol.VersionResult{Files: []string{}}, nil
	}
	if !params.DryRun {
		if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
			return protocol.VersionResult{}, err
		}
	}
	return protocol.VersionResult{Files: []string{rel}}, nil
}

// Publish publishes a package crate to its registry with cargo publish
// (CARGO_REGISTRY_TOKEN). Apps and libraries are released by their tag; an
// app with release assets gets an archive per platform instead.
func Publish(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	if params.Assets != nil {
		return BuildAssets(workspaceRoot, params)
	}
	if params.Project.Kind != protocol.ProjectPackage {
		return protocol.PublishResult{Skipped: "released by its tag"}, nil
	}
	crate := params.Project.Name
	if len(params.Project.Exports) > 0 {
		crate = params.Project.Exports[0]
	}
	// The manifests carry the release's versions while smith releases, and
	// are restored after: cargo would refuse the dirty tree.
	argv := []string{"cargo", "publish", "-p", crate, "--allow-dirty"}
	if params.DryRun {
		argv = append(argv, "--dry-run")
	}
	if err := CargoRunner(workspaceRoot, argv); err != nil {
		return protocol.PublishResult{}, err
	}
	return protocol.PublishResult{Published: !params.DryRun, Registry: "crates.io"}, nil
}

// registryClient asks crates.io; it gives up after the registry lookup timeout.
// Tests replace it.
var registryClient = &http.Client{Timeout: protocol.RegistryLookupTimeout}

// CratesAPI is crates.io's API root; tests replace it.
var CratesAPI = "https://crates.io/api/v1"

// Published asks crates.io for a crate's newest stable version.
func Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	if params.Project.Kind != protocol.ProjectPackage || len(params.Project.Exports) == 0 {
		return protocol.PublishedResult{}, nil
	}
	return NewestStable(params.Project.Exports[0]), nil
}

// NewestStable asks crates.io for a crate's newest stable version; a crate
// it does not have comes back with no version, an unreachable registry with
// Unreachable set.
func NewestStable(crate string) protocol.PublishedResult {
	result := protocol.PublishedResult{Registry: "crates.io"}
	req, _ := http.NewRequest(http.MethodGet, CratesAPI+"/crates/"+crate, nil)
	// crates.io refuses requests without a User-Agent that says who asks.
	req.Header.Set("User-Agent", "smith (https://github.com/monosmith-app/monosmith)")
	resp, err := registryClient.Do(req)
	if err != nil {
		result.Unreachable, result.Detail = true, protocol.LookupFailure(err)
		return result
	}
	defer func() { _ = resp.Body.Close() }()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return result
	case resp.StatusCode != http.StatusOK:
		result.Unreachable, result.Detail = true, resp.Status
		return result
	}
	var body struct {
		Crate struct {
			MaxStableVersion string `json:"max_stable_version"`
			MaxVersion       string `json:"max_version"`
		} `json:"crate"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		result.Unreachable, result.Detail = true, fmt.Sprintf("unreadable answer: %v", err)
		return result
	}
	result.Version = body.Crate.MaxStableVersion
	if result.Version == "" && !strings.Contains(body.Crate.MaxVersion, "-") {
		result.Version = body.Crate.MaxVersion
	}
	return result
}
