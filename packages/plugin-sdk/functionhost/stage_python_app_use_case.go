package functionhost

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// workspaceLibrariesScript is run by the venv's python: it reads the
// dependencies of pyproject.toml and, for each one that smith installed from a
// folder of the workspace (pip records that folder in direct_url.json), prints
// its folder. A registry package has no such record. It follows the
// requirements of those libraries too, and prints a JSON object of name to
// folder.
const workspaceLibrariesScript = `
import json, re, tomllib
from importlib.metadata import PackageNotFoundError, distribution
from pathlib import Path
from urllib.parse import urlparse
from urllib.request import url2pathname

def name_of(requirement):
    return re.split(r"[\s;<>=!~\[(]", requirement.strip(), maxsplit=1)[0]

def folder_of(name):
    try:
        raw = distribution(name).read_text("direct_url.json")
    except PackageNotFoundError:
        return None
    url = urlparse(json.loads(raw).get("url", "")) if raw else None
    return url2pathname(url.path) if url and url.scheme == "file" else None

found = {}
pending = list(tomllib.loads(Path("pyproject.toml").read_text(encoding="utf-8")).get("project", {}).get("dependencies", []))
while pending:
    name = name_of(pending.pop())
    if name in found:
        continue
    folder = folder_of(name)
    if folder is None:
        continue
    found[name] = folder
    pending.extend(distribution(name).requires or [])
print(json.dumps(found))
`

// StageOptions say what StagePythonApp stages.
type StageOptions struct {
	// ProjectDir is the project folder.
	ProjectDir string
	// BuildDir is the project's build folder (dist/<project dir>).
	BuildDir string
	// Host is Azure, AWS or GCP: it says which files are deployed.
	Host string
	// Python is the interpreter of the workspace venv.
	Python string
	// Tools run python; the zero value runs the real one.
	Tools Tools
}

// pythonSources are the files and folders that each host deploys. Lambda
// loads the package from the root of the zip, so src/ is flattened.
func pythonSources(host string) (sources []string, flatten bool) {
	switch host {
	case AWS:
		return []string{"src"}, true
	case Azure:
		return []string{"function_app.py", "host.json", "requirements.txt", "src"}, false
	}
	return []string{"main.py", "requirements.txt", "src"}, false
}

// skipPythonJunk leaves out what never belongs in a deployment.
func skipPythonJunk(name string, isDir bool) bool {
	return name == "__pycache__" || strings.HasSuffix(name, ".egg-info") || strings.HasSuffix(name, ".pyc")
}

// StagePythonApp copies the files that a Python function app deploys into
// <build folder>/app, with the code of the workspace libraries it needs: a
// library of the workspace is not on a registry, so the zip must carry it.
//
// Parameters: o says what to stage.
// Returns the stage folder, or an error when a deployed file is missing, the
// venv's python fails, or a library has no package to ship.
func StagePythonApp(o StageOptions) (string, error) {
	tools := o.Tools
	if tools.Output == nil {
		tools = DefaultTools()
	}
	stage := filepath.Join(o.BuildDir, "app")
	if err := os.RemoveAll(stage); err != nil {
		return "", err
	}
	if err := os.MkdirAll(stage, 0o755); err != nil {
		return "", err
	}
	sources, flatten := pythonSources(o.Host)
	for _, source := range sources {
		from := filepath.Join(o.ProjectDir, source)
		info, err := os.Stat(from)
		if err != nil {
			return "", fmt.Errorf("the %s app deploys %s, which is missing: %w", o.Host, source, err)
		}
		switch {
		case !info.IsDir():
			err = copyFile(from, filepath.Join(stage, source))
		case flatten:
			err = copyTreeSkipping(from, stage, skipPythonJunk)
		default:
			err = copyTreeSkipping(from, filepath.Join(stage, source), skipPythonJunk)
		}
		if err != nil {
			return "", err
		}
	}
	return stage, stageLibraries(tools, o, stage)
}

// stageLibraries copies the packages of the workspace libraries the app needs.
func stageLibraries(tools Tools, o StageOptions, stage string) error {
	out, err := tools.Output(o.ProjectDir, []string{o.Python, "-c", workspaceLibrariesScript})
	if err != nil {
		return fmt.Errorf("could not read the libraries of the workspace from the venv (smith install creates it): %w", err)
	}
	var libraries map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &libraries); err != nil {
		return fmt.Errorf("the venv's python answered something else than the libraries of the workspace: %w", err)
	}
	names := make([]string, 0, len(libraries))
	for name := range libraries {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		packages := libraryPackages(libraries[name], name)
		if len(packages) == 0 {
			return fmt.Errorf("%s is a library of this workspace but has no package to ship: look in %s", name, libraries[name])
		}
		for _, p := range packages {
			if err := copyTreeSkipping(p, filepath.Join(stage, filepath.Base(p)), skipPythonJunk); err != nil {
				return err
			}
		}
	}
	return nil
}

// libraryPackages are the importable packages of a library: its import name
// under src/, else every package there.
func libraryPackages(folder, name string) []string {
	wanted := strings.ToLower(strings.ReplaceAll(name, "-", "_"))
	for _, root := range []string{filepath.Join(folder, "src"), folder} {
		if _, err := os.Stat(filepath.Join(root, wanted, "__init__.py")); err == nil {
			return []string{filepath.Join(root, wanted)}
		}
	}
	inits, _ := filepath.Glob(filepath.Join(folder, "src", "*", "__init__.py"))
	slices.Sort(inits)
	packages := make([]string, len(inits))
	for i, init := range inits {
		packages[i] = filepath.Dir(init)
	}
	return packages
}

// ZipPythonApp zips the stage folder of a Python function app into
// <build folder>/<name>.zip.
//
// Parameters: buildDir is the project's build folder; name the project name.
// Returns the path of the zip, or an error when the stage is missing.
func ZipPythonApp(buildDir, name string) (string, error) {
	stage := filepath.Join(buildDir, "app")
	if _, err := os.Stat(stage); err != nil {
		return "", fmt.Errorf("nothing is staged in %s: run the build task first", stage)
	}
	zipPath := filepath.Join(buildDir, name+".zip")
	if err := os.Remove(zipPath); err != nil && !os.IsNotExist(err) {
		return "", err
	}
	return zipPath, zipFolder(stage, zipPath)
}
