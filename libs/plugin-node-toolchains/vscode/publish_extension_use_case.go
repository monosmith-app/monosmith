package vscode

import (
	"errors"
	"os"
	"path/filepath"
	"slices"

	"github.com/monosmith-app/monosmith/libs/plugin-node/releasing"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

const (
	marketplace = "https://marketplace.visualstudio.com"
	openVSX     = "https://open-vsx.org"
)

// run is the command runner; tests replace it.
var run = func(dir string, argv []string) error { return toolrun.Run(dir, argv) }

// publishExtension releases a VS Code extension that is an app: it is
// versioned by its tag, so the release's version is stamped into the manifest
// of its build, packed into a .vsix under dist/release/<app>/, and published to
// the Marketplace (with VSCE_PAT, or as the Entra ID identity the pipeline
// signed in as when VSCE_AZURE_CREDENTIAL is set) and, when OVSX_PAT is set, to
// Open VSX. Both skip a version they already have, so a re-run completes a
// release. Anything that is not an extension app is left alone.
func publishExtension(workspaceRoot string, params protocol.PublishParams) (protocol.PublishResult, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(params.Project.Root))
	if params.Project.Kind != protocol.ProjectApp {
		return protocol.PublishResult{}, nil
	}
	if _, ok := releasing.Extension(dir); !ok {
		return protocol.PublishResult{}, nil
	}
	result := protocol.PublishResult{Registry: marketplace}
	// A native extension is built on each OS: its release job only tags it, and
	// each native job publishes its own platform's build.
	native := slices.Contains(params.Project.Tags, "native")
	if native && params.Platform == "" {
		result.Skipped = "published for each platform by the native jobs"
		return result, nil
	}
	var target string
	if params.Platform != "" {
		t, err := releasing.VSCodeTarget(params.Platform)
		if err != nil {
			return result, err
		}
		target = t
	}
	ovsx := os.Getenv("OVSX_PAT") != ""
	if ovsx {
		result.Registry += ", " + openVSX
	}
	if params.DryRun {
		result.Skipped = "dry run"
		return result, nil
	}
	entra := os.Getenv("VSCE_AZURE_CREDENTIAL") != ""
	if os.Getenv("VSCE_PAT") == "" && !entra {
		return result, errors.New("set VSCE_PAT to a Marketplace token (or [release] vsce_entra = true) to publish " + params.Project.Name)
	}
	staged, err := releasing.Stage(workspaceRoot, params.Project.Root)
	if err != nil {
		return result, err
	}
	if staged == dir {
		return result, errors.New(params.Project.Name + " does not build into dist/: an extension app is packed from its build")
	}
	if err := stampVersion(staged, params.Version); err != nil {
		return result, err
	}
	vsix, err := packExtension(workspaceRoot, staged, params.Project.Name, target)
	if err != nil {
		return result, err
	}
	if err := run(staged, publishArgv(vsix, entra)); err != nil {
		return result, err
	}
	if ovsx {
		if err := run(staged, []string{"npx", "--no-install", "ovsx", "publish", vsix, "--skip-duplicate"}); err != nil {
			return result, err
		}
	}
	result.Published = true
	return result, nil
}

// stampVersion writes the release's version into the manifest of the build.
func stampVersion(staged, version string) error {
	file := filepath.Join(staged, "package.json")
	raw, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return err
	}
	doc.SetString("version", version)
	return os.WriteFile(file, doc.Bytes(), 0o644)
}

// packExtension packs the staged build into dist/release/<app>/<app>.vsix, or
// <app>-<target>.vsix for one platform.
func packExtension(workspaceRoot, staged, app, target string) (string, error) {
	drop := filepath.Join(workspaceRoot, "dist", "release", app)
	if err := os.MkdirAll(drop, 0o755); err != nil {
		return "", err
	}
	name, argv := app+".vsix", append([]string{"npx", "--no-install", "vsce", "package"}, releasing.VsceFlags...)
	if target != "" {
		name, argv = app+"-"+target+".vsix", append(argv, "--target", target)
	}
	vsix := filepath.Join(drop, name)
	return vsix, run(staged, append(argv, "--out", vsix))
}

// publishArgv is the vsce command that publishes a packed .vsix.
func publishArgv(vsix string, entra bool) []string {
	argv := []string{"npx", "--no-install", "vsce", "publish", "--packagePath", vsix}
	if entra {
		argv = append(argv, "--azure-credential")
	}
	return append(argv, "--skip-duplicate")
}
