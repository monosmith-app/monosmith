package turboworkspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"go.yaml.in/yaml/v3"
)

// ErrNotTurbo is returned when the directory has no turbo.json.
var ErrNotTurbo = errors.New("no turbo.json here: convert runs at the root of a Turborepo workspace")

// TurboFiles are the names Turborepo reads its configuration from.
var TurboFiles = []string{"turbo.json", "turbo.jsonc"}

// Read loads a Turborepo workspace: turbo.json, the root package.json,
// pnpm-workspace.yaml and every package the workspace globs match. files are
// the workspace's files (workspace-relative, forward slashes).
func Read(root string, files []string) (*Workspace, error) {
	w := &Workspace{Root: root, Files: files}
	var found bool
	for _, name := range TurboFiles {
		turbo, err := readTurbo(root, name)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		w.Turbo, found = turbo, true
		break
	}
	if !found {
		return nil, ErrNotTurbo
	}
	raw, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if err == nil {
		if err := json.Unmarshal(raw, &w.RootPackage); err != nil {
			return nil, fmt.Errorf("package.json: %w", err)
		}
		_ = json.Unmarshal(raw, &w.RootExtras)
	}
	w.TurboVersion = w.RootPackage.DevDependencies["turbo"]
	if w.TurboVersion == "" {
		w.TurboVersion = w.RootPackage.Dependencies["turbo"]
	}
	w.Manager = DetectManager(w.RootExtras.PackageManager, files)
	if raw, err := os.ReadFile(filepath.Join(root, "pnpm-workspace.yaml")); err == nil {
		w.Pnpm = &PnpmWorkspace{}
		if err := yaml.Unmarshal(raw, w.Pnpm); err != nil {
			return nil, fmt.Errorf("pnpm-workspace.yaml: %w", err)
		}
		var top map[string]any
		_ = yaml.Unmarshal(raw, &top)
		for k := range top {
			w.Pnpm.Keys = append(w.Pnpm.Keys, k)
		}
		sort.Strings(w.Pnpm.Keys)
	}
	w.Globs = w.RootPackage.WorkspaceGlobs()
	if w.Pnpm != nil && len(w.Pnpm.Packages) > 0 {
		w.Globs = w.Pnpm.Packages
	}

	for _, dir := range packageDirs(files, w.Globs) {
		p := &Package{Dir: dir}
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(dir), "package.json"))
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &p.JSON); err != nil {
			return nil, fmt.Errorf("%s/package.json: %w", dir, err)
		}
		p.Name = p.JSON.Name
		if p.Name == "" {
			p.Name = path.Base(dir)
		}
		for _, f := range files {
			if strings.HasPrefix(f, dir+"/") {
				p.Files = append(p.Files, f)
			}
		}
		for _, name := range TurboFiles {
			turbo, err := readTurbo(root, dir+"/"+name)
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			if err != nil {
				return nil, err
			}
			p.Turbo = &turbo
			break
		}
		w.Packages = append(w.Packages, p)
	}
	return w, nil
}

// packageDirs are the directories with a package.json that a workspace glob
// matches ("!" globs leave directories out), in path order.
func packageDirs(files, globs []string) []string {
	var include, exclude []string
	for _, g := range globs {
		g = strings.TrimSuffix(strings.TrimPrefix(g, "./"), "/")
		if rest, ok := strings.CutPrefix(g, "!"); ok {
			exclude = append(exclude, rest)
			continue
		}
		include = append(include, g)
	}
	var dirs []string
	for _, f := range files {
		dir, base := path.Split(f)
		dir = strings.TrimSuffix(dir, "/")
		if base != "package.json" || dir == "" || strings.Contains(dir, "node_modules") {
			continue
		}
		if matchesAny(include, dir) && !matchesAny(exclude, dir) {
			dirs = append(dirs, dir)
		}
	}
	sort.Strings(dirs)
	return dirs
}

func matchesAny(globs []string, dir string) bool {
	for _, g := range globs {
		if ok, _ := doublestar.Match(g, dir); ok {
			return true
		}
	}
	return false
}

func readTurbo(root, rel string) (TurboJSON, error) {
	var t TurboJSON
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return t, err
	}
	raw = StripJSONC(raw)
	if err := json.Unmarshal(raw, &t); err != nil {
		return t, fmt.Errorf("%s: %w", rel, err)
	}
	var top map[string]json.RawMessage
	_ = json.Unmarshal(raw, &top)
	for k := range top {
		t.Keys = append(t.Keys, k)
	}
	sort.Strings(t.Keys)
	t.File = rel
	return t, nil
}
