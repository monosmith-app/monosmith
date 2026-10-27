package dependencies

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/pelletier/go-toml/v2"
)

// The tables of a Cargo.toml that name crates.io dependencies. The workspace
// root's [workspace.dependencies] is where members share one version.
const (
	tableDependencies = "dependencies"
	tableDev          = "dev-dependencies"
	tableBuild        = "build-dependencies"
	tableWorkspace    = "workspace.dependencies"
)

// plainSpec is one version requirement sync and up understand: a bare
// version (Cargo reads it as ^), or ^, ~, = , >=, <=, < or > with one. A
// comma-separated range, a wildcard and a pre-release are left alone.
var plainSpec = regexp.MustCompile(`^(\^|~|=|>=|<=|<|>)?\s*\d+(\.\d+){0,2}$`)

// Declared lists the registry dependencies each crate's Cargo.toml declares
// and, from the root Cargo.toml, the shared [workspace.dependencies]. A
// dependency that comes from a path or a git repository, or that a crate
// inherits from the workspace (`workspace = true`), declares no version of
// its own and is left out.
func Declared(workspaceRoot string, params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	var res protocol.DeclaredResult
	files := []struct{ project, file string }{{"", "Cargo.toml"}}
	for _, p := range params.Projects {
		files = append(files, struct{ project, file string }{p.Name, path.Join(p.Root, "Cargo.toml")})
	}
	for _, f := range files {
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(f.file)))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return res, err
		}
		var m manifest
		if err := toml.Unmarshal(raw, &m); err != nil {
			return res, fmt.Errorf("%s: %w", f.file, err)
		}
		tables := map[string]map[string]any{tableDependencies: m.Dependencies, tableDev: m.DevDependencies, tableBuild: m.BuildDependencies}
		if f.project == "" {
			tables = map[string]map[string]any{tableWorkspace: m.Workspace.Dependencies}
		}
		for section, deps := range tables {
			for key, value := range deps {
				name, spec, ok := registryDependency(key, value)
				if !ok {
					continue
				}
				res.Declarations = append(res.Declarations, protocol.Declaration{
					Project: f.project, File: f.file, Name: name, Spec: spec, Section: section,
					Rewritable: plainSpec.MatchString(spec),
				})
			}
		}
	}
	sort.SliceStable(res.Declarations, func(i, j int) bool {
		a, b := res.Declarations[i], res.Declarations[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Section != b.Section {
			return a.Section < b.Section
		}
		return a.Name < b.Name
	})
	return res, nil
}

// manifest is the part of a Cargo.toml that lists dependencies.
type manifest struct {
	Dependencies      map[string]any `toml:"dependencies"`
	DevDependencies   map[string]any `toml:"dev-dependencies"`
	BuildDependencies map[string]any `toml:"build-dependencies"`
	Workspace         struct {
		Dependencies map[string]any `toml:"dependencies"`
	} `toml:"workspace"`
}

// registryDependency reads one dependency entry: its crate name (the
// `package` rename, else the key) and its version requirement. ok is false
// for entries that name no registry version.
func registryDependency(key string, value any) (name, spec string, ok bool) {
	switch v := value.(type) {
	case string:
		return key, strings.TrimSpace(v), true
	case map[string]any:
		if _, inherited := v["workspace"]; inherited {
			return "", "", false
		}
		for _, other := range []string{"path", "git"} {
			if _, has := v[other]; has {
				return "", "", false
			}
		}
		version, isString := v["version"].(string)
		if !isString {
			return "", "", false
		}
		name = key
		if pkg, renamed := v["package"].(string); renamed && pkg != "" {
			name = pkg
		}
		return name, strings.TrimSpace(version), true
	}
	return "", "", false
}
