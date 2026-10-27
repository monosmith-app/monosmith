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
	"go.yaml.in/yaml/v3"
)

// The sections of a pubspec.yaml that name pub.dev packages.
const (
	SectionDependencies = "dependencies"
	SectionDev          = "dev_dependencies"
)

// plainSpec is one constraint sync and up understand: a version, optionally
// with ^ or a comparison. `any`, a range of two comparisons and a git or path
// dependency are left alone.
var plainSpec = regexp.MustCompile(`^(\^|>=|<=|>|<|=)?\s*\d+(\.\d+){0,2}([-+][0-9A-Za-z.+-]+)?$`)

type pubspec struct {
	Dependencies    map[string]any `yaml:"dependencies"`
	DevDependencies map[string]any `yaml:"dev_dependencies"`
}

// Declared lists the pub.dev packages each package's pubspec.yaml (and the
// workspace root's, when it has one) declares. A dependency from a path, a
// git repository or the SDK declares no version of its own and is left out.
func Declared(workspaceRoot string, params protocol.DeclaredParams) (protocol.DeclaredResult, error) {
	var res protocol.DeclaredResult
	files := []struct{ project, file string }{{"", "pubspec.yaml"}}
	for _, p := range params.Projects {
		files = append(files, struct{ project, file string }{p.Name, path.Join(p.Root, "pubspec.yaml")})
	}
	for _, f := range files {
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(f.file)))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return res, err
		}
		var ps pubspec
		if err := yaml.Unmarshal(raw, &ps); err != nil {
			return res, fmt.Errorf("%s: %w", f.file, err)
		}
		for section, deps := range map[string]map[string]any{SectionDependencies: ps.Dependencies, SectionDev: ps.DevDependencies} {
			for name, value := range deps {
				spec, ok := hostedConstraint(value)
				if !ok {
					continue
				}
				res.Declarations = append(res.Declarations, protocol.Declaration{
					Project: f.project, File: f.file, Name: name, Spec: spec, Section: section, Rewritable: plainSpec.MatchString(spec),
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

// hostedConstraint is the version constraint of a dependency that comes from
// a pub server: `^1.0.0`, or the version: of an entry that names a server.
func hostedConstraint(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v), true
	case map[string]any:
		for _, other := range []string{"path", "git", "sdk"} {
			if _, has := v[other]; has {
				return "", false
			}
		}
		version, ok := v["version"].(string)
		return strings.TrimSpace(version), ok
	}
	return "", false
}
