package releasing

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// npm is the command runner; tests replace it.
var npm = struct {
	output func(dir string, argv []string) (string, error)
	run    func(dir string, argv []string) error
}{
	output: func(dir string, argv []string) (string, error) { return toolrun.Output(dir, argv) },
	run:    func(dir string, argv []string) error { return toolrun.Run(dir, argv) },
}

// Publish publishes one package with npm, or a VS Code extension to the
// Marketplace, or builds an app's release asset. A version the registry
// already has is skipped, which is what makes a re-run after a partial
// release safe.
func Publish(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	if params.Assets != nil {
		return BuildAssets(workspaceRoot, params)
	}
	// An app is deployed, never published to npm: it releases by its tag,
	// its assets and its deploy plugins (an image).
	if params.Project.Kind == protocol.ProjectApp {
		skipped := "released by its tag"
		// A manifest that is not private looks meant for npm: say how to get there.
		if raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root), "package.json")); err == nil {
			var m struct {
				Private bool `json:"private"`
			}
			if json.Unmarshal(raw, &m) == nil && !m.Private {
				skipped += `; it is not private, but an app is never sent to npm: to publish it there, make it a package ("smith": {"kind": "package"})`
			}
		}
		return protocol.PublishResult{Skipped: skipped}, nil
	}
	if len(params.Project.Exports) == 0 {
		return protocol.PublishResult{}, errors.New("package.json has no name")
	}
	name := params.Project.Exports[0]
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root))
	if _, ok := Extension(dir); ok {
		return publishExtension(workspaceRoot, params)
	}
	// npm refuses config and view inside a workspace package, so ask at the
	// root. A private registry smith.toml names for publishing is passed
	// with --registry; otherwise npm uses the .npmrc's.
	var at []string
	registry := ""
	if r := params.Project.Registries; r != nil && r.Publish != nil && r.Publish.NPM != "" {
		registry, at = r.Publish.NPM, []string{"--registry", r.Publish.NPM}
	} else {
		registry, _ = npm.output(workspaceRoot, []string{"npm", "config", "get", "registry"})
	}
	result := protocol.PublishResult{Registry: strings.TrimSpace(registry)}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	existing, _ := npm.output(workspaceRoot, append([]string{"npm", "view", name + "@" + params.Version, "version"}, at...))
	if strings.TrimSpace(existing) == params.Version {
		result.Skipped = "already published"
		return result, nil
	}
	// A project that builds into dist/ is published from there, staged.
	staged, err := Stage(workspaceRoot, params.Project.Root)
	if err != nil {
		return result, err
	}
	if staged == dir {
		err = npm.run(dir, append([]string{"npm", "publish", "--access", "public"}, at...))
	} else {
		err = npm.run(workspaceRoot, append([]string{"npm", "publish", staged, "--access", "public"}, at...))
	}
	if err != nil {
		return result, err
	}
	result.Published = true
	return result, nil
}
