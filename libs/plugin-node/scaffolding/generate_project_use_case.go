package scaffolding

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Installer runs npm install at the workspace root; tests replace it.
var Installer = func(workspaceRoot string) error {
	return toolrun.Run(workspaceRoot, []string{"npm", "install", "--no-audit", "--no-fund"})
}

type templateData struct {
	Dir, Name, Package, Stack string
	// Build is the project's build directory relative to it
	// (../../dist/apps/web): every build goes to the workspace's dist/.
	Build string
	// Up leads from the project to the workspace root (../../).
	Up string
	// TestSuffix is the workspace's test file suffix, "spec" or "test".
	TestSuffix string
}

// Generate scaffolds one plain JavaScript project.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return GenerateFrom(workspaceRoot, params, JavaScript)
}

// GenerateFrom scaffolds one Node project from a template set, adds the
// workspace's Node tooling the first time, and installs so the new package
// is linked into node_modules.
func GenerateFrom(workspaceRoot string, params protocol.GenerateParams, set TemplateSet) (protocol.GenerateResult, error) {
	data := templateData{Dir: params.Directory, Name: params.Name, Package: packageName(params.Workspace.Scope, params.Name), Stack: params.Stack, Build: protocol.BuildDirFromProject(params.Directory), Up: strings.Repeat("../", strings.Count(strings.Trim(params.Directory, "/"), "/")+1), TestSuffix: TestSuffix(workspaceRoot)}

	var files []string
	rootFiles, err := ensureWorkspace(workspaceRoot, params.Workspace.Name, set)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	files = append(files, rootFiles...)

	templates := maps.Clone(set.Slice)
	switch params.Kind {
	case protocol.ProjectApp:
		maps.Copy(templates, set.App)
	case protocol.ProjectLib:
		maps.Copy(templates, set.Lib)
	case protocol.ProjectPackage:
		maps.Copy(templates, set.Package)
	default:
		return protocol.GenerateResult{}, fmt.Errorf("unknown kind %q", params.Kind)
	}
	written, err := templating.Write(workspaceRoot, templates, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	files = append(files, written...)

	if err := Installer(workspaceRoot); err != nil {
		return protocol.GenerateResult{Files: files}, fmt.Errorf("installing dependencies: %w", err)
	}
	return protocol.GenerateResult{
		Files: files,
		Notes: []string{fmt.Sprintf("Import it from another Node project as '%s'.", data.Package)},
	}, nil
}

func packageName(scope, name string) string {
	if scope == "" {
		return name
	}
	if !strings.HasPrefix(scope, "@") {
		scope = "@" + scope
	}
	return scope + "/" + name
}

// ensureWorkspace writes the root Node files that do not exist yet. An
// existing root package.json gains the workspaces and devDependencies it
// lacks; everything already there is kept.
func ensureWorkspace(root, workspaceName string, set TemplateSet) ([]string, error) {
	var written []string
	rootFiles := maps.Clone(workspaceTemplates)
	maps.Copy(rootFiles, set.Root)
	for name, body := range rootFiles {
		p := filepath.Join(root, name)
		if _, err := os.Stat(p); err == nil {
			continue
		} else if !errors.Is(err, fs.ErrNotExist) {
			return written, err
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			return written, err
		}
		written = append(written, name)
	}
	changed, err := ensureRootManifest(filepath.Join(root, "package.json"), workspaceName, set.DevDependencies)
	if err != nil {
		return written, err
	}
	if changed {
		written = append(written, "package.json")
	}
	sort.Strings(written)
	return written, nil
}

func ensureRootManifest(path, workspaceName string, devDependencies map[string]string) (bool, error) {
	manifest := map[string]any{}
	raw, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := json.Unmarshal(raw, &manifest); err != nil {
			return false, fmt.Errorf("root package.json: %w", err)
		}
	case errors.Is(err, fs.ErrNotExist):
		manifest["name"] = strings.ToLower(workspaceName) + "-workspace"
		manifest["private"] = true
	default:
		return false, err
	}
	changed := len(raw) == 0
	if _, ok := manifest["workspaces"]; !ok {
		manifest["workspaces"] = []string{"apps/*", "libs/*", "packages/*"}
		changed = true
	}
	dev, _ := manifest["devDependencies"].(map[string]any)
	if dev == nil {
		dev = map[string]any{}
	}
	for name, version := range devDependencies {
		if _, ok := dev[name]; !ok {
			dev[name] = version
			changed = true
		}
	}
	manifest["devDependencies"] = dev
	if !changed {
		return false, nil
	}
	out, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return false, err
	}
	return true, os.WriteFile(path, append(out, '\n'), 0o644)
}
