package declaredimports

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// sdistStaging matches the `<name>-<version>` folder that `python -m build`
// stages inside the project and removes again.
var sdistStaging = regexp.MustCompile(`^[A-Za-z0-9_.]+-\d+(\.\w+)*$`)

var skippedDirs = map[string]bool{"dist": true, "build": true, "node_modules": true, "__pycache__": true, "site-packages": true}

// Check returns the imports of the Python project at project (a path relative
// to the workspace root) that its pyproject.toml does not declare where
// installing it installs them, sorted by file. Only imports that can be told
// from the standard library, the project's own modules and the declared
// distributions are reported; when a declared distribution is neither known
// nor installed in the workspace venv, nothing is reported (see doc.go).
func Check(workspaceRoot, project string) ([]UndeclaredImport, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(project))
	declared, err := readDeclarations(workspaceRoot, dir)
	if err != nil || !declared.known {
		return nil, err
	}
	files, own, err := scan(dir)
	if err != nil {
		return nil, err
	}
	imported := map[string]bool{}
	for _, f := range files {
		for _, m := range f.modules {
			imported[strings.ToLower(m)] = true
		}
	}
	installed := installedModules(workspaceRoot)
	runtime := &providers{names: declared.runtime, installed: installed, imported: imported}
	dev := &providers{names: declared.dev, installed: installed, imported: imported}
	if runtime.opaque() {
		return nil, nil // one of them may provide any module
	}
	var out []UndeclaredImport
	for _, f := range files {
		devFile := DevFile(f.rel)
		if devFile && dev.opaque() {
			continue
		}
		for _, m := range f.modules {
			if stdlib[m] || own[m] || runtime.provide(m) || (devFile && dev.provide(m)) {
				continue
			}
			out = append(out, UndeclaredImport{File: f.rel, Module: m, DevOnly: dev.provide(m)})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].File < out[j].File })
	return out, nil
}

type scannedFile struct {
	rel     string
	modules []string
}

// scan reads the project's Python files: each file's imports, deduplicated,
// and every module name the project itself provides (a folder, or a file's
// stem, at any depth: pytest and scripts put their own folder on sys.path).
func scan(dir string) ([]scannedFile, map[string]bool, error) {
	var files []scannedFile
	own := map[string]bool{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if errors.Is(err, fs.ErrNotExist) {
			return nil // a build folder removed while the walk ran
		}
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if p != dir && (skippedDirs[name] || strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".egg-info") || sdistStaging.MatchString(name)) {
				return filepath.SkipDir
			}
			own[name] = true
			return nil
		}
		if path.Ext(name) != ".py" {
			return nil
		}
		own[strings.TrimSuffix(name, ".py")] = true
		src, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		seen := map[string]bool{}
		f := scannedFile{rel: filepath.ToSlash(rel)}
		for _, m := range FileImports(string(src)) {
			if !seen[m] {
				seen[m] = true
				f.modules = append(f.modules, m)
			}
		}
		files = append(files, f)
		return nil
	})
	return files, own, err
}

// providers answers which modules a set of declared distributions provide.
type providers struct {
	names     map[string]bool
	installed map[string][]string
	imported  map[string]bool
}

func (p *providers) provide(module string) bool {
	for dist := range p.names {
		if p.provides(dist, module) {
			return true
		}
	}
	return false
}

func (p *providers) provides(dist, module string) bool {
	if strings.EqualFold(strings.ReplaceAll(dist, "-", "_"), module) {
		return true
	}
	for _, m := range knownImports[dist] {
		if m == module {
			return true
		}
	}
	for _, m := range p.installed[dist] {
		if m == module {
			return true
		}
	}
	return false
}

// opaque reports whether a distribution is declared whose modules are
// unknown: not in the table, not installed, and not named by an import the
// project makes (which would explain it).
func (p *providers) opaque() bool {
	for dist := range p.names {
		_, known := knownImports[dist]
		_, installed := p.installed[dist]
		if !known && !installed && !p.imported[strings.ReplaceAll(dist, "-", "_")] {
			return true
		}
	}
	return false
}
