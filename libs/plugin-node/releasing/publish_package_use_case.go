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
	if len(params.Finalize) > 0 {
		return Finalize(workspaceRoot, params)
	}
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
			// One quick probe, never a wait: the host moves the dist-tag after
			// every project has published (see Finalize).
			if out := awaitPlatforms(workspaceRoot, at, 0, platforms, params.Version); !out.Ready {
				// A launcher whose platform packages cannot be fetched yet would install
				// and fail: keep the dist-tag where it is; the host moves it at the end.
				tag = pendingTag
				result.PendingMoves = append(result.PendingMoves, protocol.TagMove{Package: name, Version: params.Version, Tag: cmp.Or(params.Tag, "latest"), Awaits: platforms})
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

// Finalize is the last phase of a release: for each move it waits, for up to
// params.PropagationWait, until the packages the move awaits are fetchable the
// way a person fetches them, then moves the dist-tag. A move whose packages do
// not appear in time stays in PendingMoves with a note naming the command that
// moves it by hand.
//
// Parameters: workspaceRoot is the workspace; params carries the moves, the
// project (for its registry) and the wait.
// Returns: the moves done (Moved) and the ones still waiting (PendingMoves).
// Errors: the wait is not a duration, or npm refuses to move a tag.
func Finalize(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	var result protocol.PublishResult
	wait, err := npmpropagation.ParseWait(params.PropagationWait)
	if err != nil {
		return result, err
	}
	var at []string
	if r := params.Project.Registries; r != nil && r.Publish != nil && r.Publish.NPM != "" {
		at = []string{"--registry", r.Publish.NPM}
	}
	for _, move := range params.Finalize {
		out := awaitPlatforms(workspaceRoot, at, wait, move.Awaits, move.Version)
		if !out.Ready {
			result.PendingMoves = append(result.PendingMoves, move)
			result.Notes = append(result.Notes, notFetchableNote(move, out))
			continue
		}
		if err := npm.run(workspaceRoot, append([]string{"npm", "dist-tag", "add", move.Package + "@" + move.Version, move.Tag}, at...)); err != nil {
			return result, fmt.Errorf("moving the dist-tag %s of %s: %w", move.Tag, move.Package, err)
		}
		// The pending tag has done its job; a leftover would make a re-run look for work.
		_ = npm.run(workspaceRoot, append([]string{"npm", "dist-tag", "rm", move.Package, pendingTag}, at...))
		result.Moved = append(result.Moved, move.Package+"@"+move.Version+" "+move.Tag)
	}
	return result, nil
}

// notFetchableNote says why a dist-tag did not move and how to move it.
func notFetchableNote(move protocol.TagMove, out npmpropagation.Outcome) string {
	missing := "the platform packages"
	if len(out.Missing) > 0 {
		var named []string
		for _, n := range out.Missing {
			named = append(named, n+"@"+move.Version)
		}
		missing = strings.Join(named, ", ")
	}
	return fmt.Sprintf("the dist-tag %q of %s was not moved: npm said it published %s but the registry still does not serve it after %s. "+
		"Run the release again: it republishes what is missing. Or, when npm serves it, run: %s",
		move.Tag, move.Package, missing, out.Waited.Round(time.Second), move.Command())
}
