package releasing

import (
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
		registry, _, _ = capture(workspaceRoot, []string{"npm", "config", "get", "registry"})
	}
	result := protocol.PublishedResult{Registry: strings.TrimSpace(registry)}
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
	return result, nil
}

func firstLine(text, fallback string) string {
	for _, line := range strings.Split(text, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return fallback
}
