package pruning

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/releasing"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Task is the name of the task that prunes an app.
const Task = "prune"

// Request names the app and where its deploy output goes.
type Request struct {
	WorkspaceRoot string
	// Root is the app's directory, workspace-relative.
	Root string
	// Out is the directory the output is written into, workspace-relative:
	// the app's build directory, which must exist. Default: the build
	// directory smith builds into (dist/<app dir>).
	Out string
}

// Result is what the output holds.
type Result struct {
	// Out is the output directory, workspace-relative.
	Out string
	// Modules are the workspace packages copied into workspace_modules/.
	Modules []string
}

// npm runs npm; tests replace it.
var npm = func(dir string, args ...string) error {
	return toolrun.Run(dir, append([]string{"npm"}, args...))
}

// Prune writes the app's deploy output into its build directory: a
// package.json whose workspace dependencies point at copies in
// workspace_modules/ (each one as it would be published: its build with its
// package.json), and a lockfile of the production tree at the versions the
// workspace installed. The directory then installs on its own with
// npm ci --omit=dev.
func Prune(req Request) (Result, error) {
	out := req.Out
	if out == "" {
		out = protocol.ProjectBuildDir(req.Root)
	}
	res := Result{Out: out}
	ws := req.WorkspaceRoot
	outDir := filepath.Join(ws, filepath.FromSlash(out))
	if _, err := os.Stat(outDir); errors.Is(err, fs.ErrNotExist) {
		return res, fmt.Errorf("%s does not exist: build %s first (smith run build -p %s)", out, req.Root, path.Base(req.Root))
	}
	lock, err := os.ReadFile(filepath.Join(ws, "package-lock.json"))
	if err != nil {
		return res, fmt.Errorf("pruning needs the workspace's package-lock.json (run npm install): %w", err)
	}
	workspace, err := WorkspacePackages(lock)
	if err != nil {
		return res, fmt.Errorf("package-lock.json: %w", err)
	}
	isWorkspace := func(name string) bool { _, ok := workspace[name]; return ok }
	raw, err := os.ReadFile(filepath.Join(ws, filepath.FromSlash(req.Root), "package.json"))
	if err != nil {
		return res, err
	}
	if out == protocol.ProjectBuildDir(req.Root) {
		// The package.json is read from the build directory now.
		raw = []byte(releasing.RelativeToBuild(string(raw), req.Root))
	}
	manifest, queue, err := PrunedManifest(raw, isWorkspace, func(name string) string { return "file:./" + ModulesDir + "/" + name })
	if err != nil {
		return res, fmt.Errorf("%s/package.json: %w", req.Root, err)
	}
	modulesDir := filepath.Join(outDir, ModulesDir)
	if err := os.RemoveAll(modulesDir); err != nil {
		return res, err
	}
	if err := os.MkdirAll(modulesDir, 0o755); err != nil {
		return res, err
	}
	copied := map[string]string{}
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if _, done := copied[name]; done {
			continue
		}
		copied[name] = workspace[name]
		deps, err := copyModule(ws, workspace[name], filepath.Join(modulesDir, filepath.FromSlash(name)), name, isWorkspace)
		if err != nil {
			return res, fmt.Errorf("copying %s: %w", name, err)
		}
		queue = append(queue, deps...)
		res.Modules = append(res.Modules, name)
	}
	if err := os.WriteFile(filepath.Join(outDir, "package.json"), manifest, 0o644); err != nil {
		return res, err
	}
	var id struct{ Name, Version string }
	if err := json.Unmarshal(manifest, &id); err != nil {
		return res, err
	}
	seed, err := SeedLockfile(lock, id.Name, id.Version, req.Root, copied)
	if err != nil {
		return res, err
	}
	if err := os.WriteFile(filepath.Join(outDir, "package-lock.json"), seed, 0o644); err != nil {
		return res, err
	}
	return res, installLockfile(ws, outDir)
}

// copyModule copies a workspace package into the output as it would be
// published (its build, with its package.json), and points its own workspace
// dependencies at their sibling copies. It returns those dependencies.
func copyModule(ws, dir, dest, name string, isWorkspace func(string) bool) ([]string, error) {
	staged, err := releasing.Stage(ws, dir)
	if err != nil {
		return nil, err
	}
	if err := copyTree(staged, dest); err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(filepath.Join(dest, "package.json"))
	if err != nil {
		return nil, err
	}
	from := path.Join(ModulesDir, name)
	manifest, deps, err := PrunedManifest(raw, isWorkspace, func(sibling string) string {
		rel, _ := filepath.Rel(filepath.FromSlash(from), filepath.FromSlash(path.Join(ModulesDir, sibling)))
		return "file:" + filepath.ToSlash(rel)
	})
	if err != nil {
		return nil, err
	}
	return deps, os.WriteFile(filepath.Join(dest, "package.json"), manifest, 0o644)
}

// copyTree copies a directory, leaving out node_modules.
func copyTree(from, to string) error {
	if err := os.RemoveAll(to); err != nil {
		return err
	}
	return filepath.WalkDir(from, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == "node_modules" {
			return filepath.SkipDir
		}
		rel, err := filepath.Rel(from, p)
		if err != nil {
			return err
		}
		target := filepath.Join(to, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !d.Type().IsRegular() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		content, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, content, info.Mode().Perm())
	})
}

// installLockfile has npm prune the seeded lockfile to the production tree
// and record the copied packages, without installing anything. The
// workspace's .npmrc (a private registry, say) applies while it runs, and is
// not left in the output.
func installLockfile(ws, outDir string) error {
	npmrc := filepath.Join(outDir, ".npmrc")
	if _, err := os.Stat(npmrc); errors.Is(err, fs.ErrNotExist) {
		if content, err := os.ReadFile(filepath.Join(ws, ".npmrc")); err == nil {
			if err := os.WriteFile(npmrc, content, 0o600); err != nil {
				return err
			}
			defer func() { _ = os.Remove(npmrc) }()
		}
	}
	if err := npm(outDir, "install", "--package-lock-only", "--ignore-scripts", "--no-audit", "--no-fund"); err != nil {
		return fmt.Errorf("npm install --package-lock-only in %s: %w", strings.TrimPrefix(filepath.ToSlash(outDir), filepath.ToSlash(ws)+"/"), err)
	}
	return nil
}
