package releasing

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// WriteVersion records a release's version in the composer.json.
func WriteVersion(workspaceRoot string, params protocol.VersionParams) (protocol.VersionResult, error) {
	rel := path.Join(params.Project.Root, "composer.json")
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

// Packagist is Packagist's package metadata; tests replace it.
var Packagist = "https://repo.packagist.org"

// Publish has nothing to upload: Packagist reads a package's versions from
// its repository's tags, which the release pushes.
func Publish(params protocol.PublishParams) (protocol.PublishResult, error) {
	if params.Project.Kind != protocol.ProjectPackage {
		return protocol.PublishResult{Skipped: "released by its tag"}, nil
	}
	return protocol.PublishResult{Registry: "packagist.org", Skipped: "Packagist publishes from the repository's tags"}, nil
}

// Published asks Packagist for a package's newest stable version.
func Published(params protocol.PublishedParams) (protocol.PublishedResult, error) {
	if params.Project.Kind != protocol.ProjectPackage || len(params.Project.Exports) == 0 {
		return protocol.PublishedResult{}, nil
	}
	name := params.Project.Exports[0]
	result := protocol.PublishedResult{Registry: "packagist.org"}
	resp, err := (&http.Client{Timeout: protocol.RegistryLookupTimeout}).Get(Packagist + "/p2/" + name + ".json")
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
		Packages map[string][]struct{ Version string } `json:"packages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		result.Unreachable, result.Detail = true, protocol.LookupFailure(err)
		return result, nil
	}
	for _, v := range body.Packages[name] {
		version := strings.TrimPrefix(v.Version, "v")
		if strings.ContainsAny(version, "-@") || strings.HasPrefix(version, "dev") {
			continue
		}
		if result.Version == "" || newer(version, result.Version) {
			result.Version = version
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
