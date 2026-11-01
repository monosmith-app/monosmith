package discovery

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

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var requireCall = regexp.MustCompile(`\brequire\s*\(?\s*["']([A-Za-z_][\w-]*)`)

// DiscoverWithDeclared is Discover plus the projects smith.toml declares: a
// folder of Lua modules that belongs to the rock whose rockspec is above it,
// with no rockspec of its own. A declaration is refused with an error that
// names it when its folder is missing, has a rockspec, has none above it, or
// asks for the kind package.
func DiscoverWithDeclared(workspaceRoot string, manifests []string, declared []protocol.DeclaredProject) (protocol.DiscoverResult, error) {
	result, err := Discover(workspaceRoot, manifests)
	if err != nil || len(declared) == 0 {
		return result, err
	}
	for _, d := range declared {
		project, err := describeDeclared(workspaceRoot, d)
		if err != nil {
			return result, fmt.Errorf("[projects.%s] at %s: %w", d.Name, d.Root, err)
		}
		result.Projects = append(result.Projects, project)
	}
	return result, nil
}

// describeDeclared describes one declared folder as a project of the rock
// whose rockspec is above it. The folder's name is the module it exports, and
// the top-level modules its sources require are its imports. luacheck and
// busted run in the folder, and the rockspec above is an input of both, since
// the folder does not hold it. There is no build: luarocks installs a rock
// whole, from its rockspec.
func describeDeclared(workspaceRoot string, d protocol.DeclaredProject) (protocol.Project, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(d.Root))
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return protocol.Project{}, fmt.Errorf("the folder does not exist")
	}
	if found, _ := filepath.Glob(filepath.Join(dir, ManifestPattern)); len(found) > 0 {
		return protocol.Project{}, fmt.Errorf("the folder has a rockspec of its own, so it is a project smith finds itself: remove the declaration")
	}
	manifest := ""
	for up := path.Dir(d.Root); ; up = path.Dir(up) {
		found, _ := filepath.Glob(filepath.Join(workspaceRoot, filepath.FromSlash(up), ManifestPattern))
		if len(found) > 0 {
			manifest = path.Join(up, filepath.Base(found[0]))
			break
		}
		if up == "." {
			break
		}
	}
	if manifest == "" {
		return protocol.Project{}, fmt.Errorf("no rockspec is above the folder: add one, or give the folder a rockspec of its own")
	}
	kind := d.Kind
	if kind == protocol.ProjectPackage {
		return protocol.Project{}, fmt.Errorf("a folder of a rock cannot be released on its own (a rockspec has one version): declare it as a lib or an app")
	}
	if kind == "" {
		kind = projectkind.Resolve("", d.Root, protocol.ProjectLib)
		if kind == protocol.ProjectPackage {
			kind = protocol.ProjectLib
		}
	}
	module := path.Base(d.Root)
	imports, err := scanRequires(dir)
	if err != nil {
		return protocol.Project{}, err
	}
	inputs := append([]string{}, inputs...)
	inputs = append(inputs, "{workspaceRoot}/"+manifest)
	task := func(command []string, description string) protocol.Task {
		return protocol.Task{Command: command, Inputs: inputs, Cache: true, Description: description}
	}
	tasks := map[string]protocol.Task{"lint": task([]string{"luacheck", "."}, "luacheck (.luacheckrc's rules)")}
	if info, err := os.Stat(filepath.Join(dir, "spec")); err == nil && info.IsDir() {
		tasks["test"] = task([]string{"busted"}, "busted (spec/)")
	}
	return protocol.Project{
		Name:     d.Name,
		Root:     d.Root,
		Language: "lua",
		Kind:     kind,
		Exports:  []string{module},
		Imports:  imports,
		Tasks:    tasks,
	}, nil
}

// scanRequires returns the top-level module names the .lua files under dir
// require, sorted and unique.
func scanRequires(dir string) ([]string, error) {
	seen := map[string]bool{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) && p != dir {
				return nil // removed while the walk ran
			}
			return err
		}
		if d.IsDir() {
			if p != dir && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(p) != ".lua" {
			return nil
		}
		src, err := os.ReadFile(p)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		for _, m := range requireCall.FindAllStringSubmatch(string(src), -1) {
			seen[m[1]] = true
		}
		return nil
	})
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out, err
}
