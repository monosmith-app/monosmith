package scaffolding

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-node/modulealiases"
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
	// BuildLocal is true when the shape asks for the build in the project's own
	// dist folder; Build is then "dist".
	BuildLocal bool
	// BuildConfig is the tsconfig file the build uses (tsconfig.build.json, or
	// tsconfig.lib.json when the siblings have one).
	BuildConfig string
	// TsconfigSpec is true when the siblings have a tsconfig.spec.json.
	TsconfigSpec bool
	// Up leads from the project to the workspace root (../../).
	Up string
	// TestSuffix is the workspace's test file suffix, "spec" or "test".
	TestSuffix string
	// TSPaths is the tsconfig "paths" member for the module.aliases fact,
	// with its leading comma, or empty for no aliases. Put it after the last
	// compilerOptions member.
	TSPaths string
	// ViteAlias is the Vite resolve.alias member for the module.aliases fact,
	// or empty for no aliases.
	ViteAlias string
	// Shape is the dialect the project follows (GenerateParams.Shape); a
	// template reads a fact with (index .Shape "build.tool"). Nil when there is
	// none, and index then gives "".
	Shape map[string]string
	// Slots, Scripts and Dependencies come from the template set (see
	// TemplateSet).
	Slots        map[string][]string
	Scripts      []Pair
	Dependencies []Pair
}

// Generate scaffolds one plain JavaScript project.
func Generate(workspaceRoot string, params protocol.GenerateParams) (protocol.GenerateResult, error) {
	return GenerateFrom(workspaceRoot, params, JavaScript)
}

// GenerateFrom scaffolds one Node project from a template set, adds the
// workspace's Node tooling the first time. It reports InstallPending: the
// host installs once, so the new package is linked into node_modules.
func GenerateFrom(workspaceRoot string, params protocol.GenerateParams, set TemplateSet) (protocol.GenerateResult, error) {
	data := newTemplateData(workspaceRoot, params, set)

	var files []string
	rootFiles, err := ensureWorkspace(workspaceRoot, params.Workspace.Name, set)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	files = append(files, rootFiles...)

	templates, err := kindTemplates(params.Kind, set)
	if err != nil {
		return protocol.GenerateResult{}, err
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

// newTemplateData is what the templates of a project read.
func newTemplateData(workspaceRoot string, params protocol.GenerateParams, set TemplateSet) templateData {
	data := templateData{Dir: params.Directory, Name: params.Name, Package: packageName(params.Workspace.Scope, params.Name), Stack: params.Stack, Build: protocol.BuildDirFromProject(params.Directory), Up: strings.Repeat("../", strings.Count(strings.Trim(params.Directory, "/"), "/")+1), TestSuffix: TestSuffix(workspaceRoot), Shape: params.Shape, Slots: set.Slots, Scripts: pairs(set.Scripts), Dependencies: pairs(set.Dependencies)}
	if params.Shape[discovery.ShapeBuildOutput] == discovery.ShapeOutputProject {
		data.Build, data.BuildLocal = "dist", true
	}
	data.BuildConfig, data.TsconfigSpec = TsconfigSet(params.Shape)
	aliases := modulealiases.FromFacts(params.Facts)
	data.TSPaths, data.ViteAlias = aliases.TSConfigPaths(), aliases.ViteResolveAlias()
	return data
}

// kindTemplates are the templates of a project of the kind: the set's shared
// ones and those of the kind. It returns an error for an unknown kind.
func kindTemplates(kind protocol.ProjectKind, set TemplateSet) (map[string]string, error) {
	templates := maps.Clone(set.Slice)
	switch kind {
	case protocol.ProjectApp:
		maps.Copy(templates, set.App)
	case protocol.ProjectLib:
		maps.Copy(templates, set.Lib)
	case protocol.ProjectPackage:
		maps.Copy(templates, set.Package)
	default:
		return nil, fmt.Errorf("unknown kind %q", kind)
	}
	return templates, nil
}

// pairs sorts a map by name, with the values escaped for a JSON string.
func pairs(m map[string]string) []Pair {
	out := make([]Pair, 0, len(m))
	for _, name := range slices.Sorted(maps.Keys(m)) {
		// HTML escaping would write && as &: a script is shell text, not HTML.
		var quoted bytes.Buffer
		enc := json.NewEncoder(&quoted)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(m[name])
		text := strings.TrimSuffix(quoted.String(), "\n")
		out = append(out, Pair{Name: name, Value: text[1 : len(text)-1]})
	}
	return out
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
