package scaffolding

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/lineendings"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/templating"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Installer runs npm install at the workspace root; tests replace it.
var Installer = func(workspaceRoot string) error {
	return toolrun.Quiet(workspaceRoot, []string{"npm", "install", "--no-audit", "--no-fund"})
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
// workspace's Node tooling the first time. It reports InstallPending: the
// host installs once, so the new package is linked into node_modules.
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

	// The host installs once, after every plugin of the stack has had its say.
	return protocol.GenerateResult{
		Files:          files,
		Notes:          []string{fmt.Sprintf("Import it from another Node project as '%s'.", data.Package)},
		InstallPending: true,
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
		if err := os.WriteFile(p, []byte(lineendings.Resolve(root, name, body)), 0o644); err != nil {
			return written, err
		}
		written = append(written, name)
	}
	changed, err := ensureRootManifest(filepath.Join(root, "package.json"), workspaceName, set.DevDependencies, set.Overrides)
	if err != nil {
		return written, err
	}
	if changed {
		written = append(written, "package.json")
	}
	sort.Strings(written)
	return written, nil
}

func ensureRootManifest(path, workspaceName string, devDependencies, overrides map[string]string) (bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	out, changed, err := completeRootManifest(raw, workspaceName, devDependencies, overrides)
	if err != nil {
		return false, fmt.Errorf("root package.json: %w", err)
	}
	if !changed {
		return false, nil
	}
	text := lineendings.Resolve(filepath.Dir(path), filepath.Base(path), string(out))
	return true, os.WriteFile(path, []byte(text), 0o644)
}
