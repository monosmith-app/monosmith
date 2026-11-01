package releasing

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// capture runs npm and returns stdout and stderr; tests replace it.
var capture = func(dir string, argv []string) (string, string, error) {
	return toolrun.Capture(dir, argv)
}

// Published asks the registry .npmrc points at for the newest version of a
// package, or the Marketplace for a VS Code extension's. A package the registry does not know (E404) is not published yet;
// any other failure means the registry could not be asked.
func Published(workspaceRoot string, params protocol.PublishedParams) (protocol.PublishedResult, error) {
	if len(params.Project.Exports) == 0 {
		return protocol.PublishedResult{}, nil
	}
	if m, ok := Extension(filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root))); ok {
		return publishedExtension(workspaceRoot, m), nil
	}
	name := params.Project.Exports[0]
	// The registry the project is published to, when smith.toml names one.
	var at []string
	registry := ""
	if r := params.Project.Registries; r != nil && r.Publish != nil && r.Publish.NPM != "" {
		registry, at = r.Publish.NPM, []string{"--registry", r.Publish.NPM}
	} else {
		registry = configuredRegistry(workspaceRoot)
	}
	result := protocol.PublishedResult{Registry: strings.TrimSpace(registry)}
	// A bare name on the public registry belongs to whoever published it first,
	// not to this repository: the version found there says nothing about it.
	if !ownsName(workspaceRoot, params.Project, name, result.Registry, len(at) > 0) {
		return result, nil
	}
	return view(workspaceRoot, name, at, result), nil
}

// Latest asks the registry .npmrc points at for the newest version of any
// package a project depends on. Unlike Published it does not ask who owns the
// name: a dependency belongs to whoever published it.
func Latest(workspaceRoot, name string) protocol.PublishedResult {
	return view(workspaceRoot, name, nil, protocol.PublishedResult{Registry: strings.TrimSpace(configuredRegistry(workspaceRoot))})
}

// configuredRegistry is the registry npm is configured with at the workspace
// root, asked once per root for a short while.
func configuredRegistry(workspaceRoot string) string {
	return registries.configured(workspaceRoot, func() string {
		registry, _, _ := capture(workspaceRoot, []string{"npm", "config", "get", "registry"})
		return registry
	})
}

// view runs npm view for a package's version and records the answer in result.
func view(workspaceRoot, name string, at []string, result protocol.PublishedResult) protocol.PublishedResult {
	// npm refuses view inside a workspace package, so ask at the root.
	out, errOut, err := capture(workspaceRoot, append([]string{"npm", "view", name, "version", "--fetch-retries=0", "--fetch-timeout=15000"}, at...))
	switch {
	case err == nil:
		result.Version = strings.TrimSpace(out)
	case strings.Contains(errOut, "E404") || strings.Contains(errOut, "404 Not Found"):
	default:
		result.Unreachable = true
		result.Detail = firstLine(errOut, err.Error())
	}
	return result
}

// ownsName reports whether a registry's answer about name is about this
// project's package. It is when the package is public and either scoped (the
// scope is the owner's), or the registry is not the public one (a private
// feed, or one smith.toml names), where a bare name is the repository's own.
func ownsName(workspaceRoot string, p protocol.Project, name, registry string, explicit bool) bool {
	var manifest struct {
		Private bool `json:"private"`
	}
	if raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "package.json")); err == nil && json.Unmarshal(raw, &manifest) == nil && manifest.Private {
		return false
	}
	if strings.HasPrefix(name, "@") || explicit {
		return true
	}
	return !strings.Contains(registry, "registry.npmjs.org") && !strings.Contains(registry, "registry.yarnpkg.com")
}

func firstLine(text, fallback string) string {
	for _, line := range strings.Split(text, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return fallback
}
