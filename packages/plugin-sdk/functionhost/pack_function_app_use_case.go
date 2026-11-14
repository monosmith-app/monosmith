package functionhost

import (
	"fmt"
	"os"
	"path/filepath"
)

// PackOptions say what PackFunctionApp packs.
type PackOptions struct {
	// ProjectDir is the project folder.
	ProjectDir string
	// BuildDir is the folder the build writes: the compiled code with the
	// manifest and host.json beside it (AssembleFunctionApp).
	BuildDir string
	// Name is the project name: the zip is <Name>.zip in BuildDir.
	Name string
	// WithDependencies installs the production dependencies into the zip, for
	// a host that does not install them itself (Azure Functions).
	WithDependencies bool
	// Tools run npm; the zero value runs the real one.
	Tools Tools
}

// PackFunctionApp builds a function app, then zips it: the compiled code, the
// manifest and, when the host does not install them, the production
// dependencies (installed with the workspace's .npmrc, so a private feed works; the
// file never reaches the zip). A library of this workspace is not on the registry: it is
// built and copied under vendor/, and the manifest of the stage points at the
// copy, so that npm installs it from there.
//
// Parameters: o says what to pack.
// Returns the path of the zip, or an error when a build, a copy, an npm
// command or the zip fails.
func PackFunctionApp(o PackOptions) (string, error) {
	tools := o.Tools
	if tools.Run == nil {
		tools = DefaultTools()
	}
	if err := tools.Run(o.ProjectDir, []string{"npm", "run", "build"}); err != nil {
		return "", err
	}
	stage := filepath.Join(o.BuildDir, "stage")
	zipPath := filepath.Join(o.BuildDir, o.Name+".zip")
	if err := os.RemoveAll(stage); err != nil {
		return "", err
	}
	if err := os.Remove(zipPath); err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if err := os.MkdirAll(stage, 0o755); err != nil {
		return "", err
	}
	for _, item := range []string{"dist", "package.json", "host.json"} {
		from := filepath.Join(o.BuildDir, item)
		info, err := os.Stat(from)
		if err != nil {
			continue
		}
		if info.IsDir() {
			err = copyTree(from, filepath.Join(stage, item))
		} else {
			err = copyFile(from, filepath.Join(stage, item))
		}
		if err != nil {
			return "", err
		}
	}
	staged, err := readManifest(filepath.Join(stage, "package.json"))
	if err != nil {
		return "", fmt.Errorf("the build left no package.json in %s: %w", o.BuildDir, err)
	}
	if err := vendorLibraries(tools, o.ProjectDir, stage, staged); err != nil {
		return "", err
	}
	if deps := stringMap(staged["dependencies"]); o.WithDependencies && len(deps) > 0 {
		argv := []string{"npm", "install", "--omit=dev", "--ignore-scripts", "--no-audit", "--no-fund", "--workspaces=false"}
		// The stage is outside the workspace: give npm the .npmrc that maps a scope to a private feed.
		removeNpmrc, err := stageNpmrc(o.ProjectDir, stage)
		if err != nil {
			return "", err
		}
		err = tools.Run(stage, argv)
		removeNpmrc()
		if err != nil {
			return "", err
		}
		// The platform that runs the zip has node_modules; the copies of the libraries are not needed.
		if err := os.RemoveAll(filepath.Join(stage, "vendor")); err != nil {
			return "", err
		}
	}
	if err := zipFolder(stage, zipPath); err != nil {
		return "", err
	}
	return zipPath, os.RemoveAll(stage)
}

// vendorLibraries copies the libraries of the workspace that the app needs
// into the stage, and points the manifest of the stage at the copies. A
// library of a library is overridden, a direct one is pointed at.
func vendorLibraries(tools Tools, projectDir, stage string, staged map[string]any) error {
	libraries := map[string]string{}
	if err := workspaceLibraries(projectDir, stringMap(staged["dependencies"]), libraries); err != nil {
		return err
	}
	if len(libraries) == 0 {
		return nil
	}
	overrides := stringMap(staged["overrides"])
	if overrides == nil {
		overrides = map[string]any{}
	}
	staged["overrides"] = overrides
	deps := stringMap(staged["dependencies"])
	for name, folder := range libraries {
		tarball, err := vendorLibrary(tools, name, folder, filepath.Join(stage, "vendor"))
		if err != nil {
			return err
		}
		spec := "file:./vendor/" + tarball
		if _, direct := deps[name]; direct {
			deps[name] = spec
		} else {
			overrides[name] = spec
		}
	}
	return writeManifest(filepath.Join(stage, "package.json"), staged)
}
