package releasing

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/npmpropagation"
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
	// An extension app has no archives: the vscode plugin packs and publishes
	// its .vsix.
	if params.Project.Kind == protocol.ProjectApp {
		if _, ok := Extension(filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root))); ok {
			return protocol.PublishResult{Skipped: "an extension app is published by the vscode plugin (stack vscode-extension)"}, nil
		}
	}
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
		registry = registries.configured(workspaceRoot, func() string {
			out, _ := npm.output(workspaceRoot, []string{"npm", "config", "get", "registry"})
			return out
		})
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
	// A release channel publishes under its dist-tag, so latest stays put (npm
	// refuses a prerelease without one).
	tag := params.Tag
	if manifest, readErr := os.ReadFile(filepath.Join(staged, "package.json")); readErr == nil {
		if platforms := SameReleaseOptionalDependencies(manifest, params.Version); len(platforms) > 0 {
			wait, werr := npmpropagation.ParseWait(params.PropagationWait)
			if werr != nil {
				return result, werr
			}
			if out := awaitPlatforms(workspaceRoot, at, wait, platforms, params.Version); !out.Ready {
				// A launcher whose platform packages cannot be fetched yet would install
				// and fail: keep the dist-tag where it is, and say how to move it.
				moved := cmp.Or(params.Tag, "latest")
				tag = pendingTag
				command := fmt.Sprintf("npm dist-tag add %s@%s %s", name, params.Version, moved)
				result.PendingTags = append(result.PendingTags, command)
				result.Notes = append(result.Notes, notFetchableNote(moved, out, len(platforms), command))
			}
		}
	}
	// A release channel publishes under its dist-tag; a package that must wait
	// publishes under "pending". Either way latest stays put.
	if tag != "" {
		at = append(at, "--tag", tag)
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

// pendingTag is the dist-tag a package takes when the channel's own tag has to
// wait for its platform packages to appear on the registry.
const pendingTag = "pending"

// awaitPlatforms waits, for up to wait, until every package is fetchable the
// way a person fetches it (found by npm view and its tarball answers), on
// three probes 30 s apart. Tests replace it with a fake registry.
var awaitPlatforms = func(workspaceRoot string, at []string, wait time.Duration, names []string, version string) npmpropagation.Outcome {
	reg := npmpropagation.NPM{Dir: workspaceRoot, Args: at, Output: npm.output}
	cfg := npmpropagation.Config{Max: wait, Progress: func(line string) { fmt.Fprintln(os.Stderr, line) }}
	return npmpropagation.Await(reg, npmpropagation.RealClock{}, cfg, names, version)
}

// notFetchableNote says why a dist-tag did not move and how to move it.
func notFetchableNote(tag string, out npmpropagation.Outcome, total int, command string) string {
	why := fmt.Sprintf("%d platform packages were not confirmed on the registry in time", total)
	if len(out.Missing) > 0 {
		why = strings.Join(out.Missing, ", ") + " could not be fetched from the registry in time"
	}
	return fmt.Sprintf("the dist-tag %q was not moved: %s (waited %s). Published as %q; when npm serves them, run: %s",
		tag, why, out.Waited.Round(time.Second), pendingTag, command)
}