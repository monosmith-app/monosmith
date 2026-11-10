package adoptsplit

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// SplitNode cuts the files of a Node project that moved into req.Dir. The
// project's package.json loses the devDependencies that go to the workspace
// (and the ESLint stack the shared lint config replaces); the root
// package.json, written here when the root has none, gains the first and the
// "workspaces" entry. tsconfig.json is shared through a root
// tsconfig.base.json. A legacy .eslintrc at the root is kept as
// .pre-smith. Nothing is staged or committed.
//
// Parameters: req is the move; Hoist and Drop are the settled dependency lists.
// Returns: the notes for the person and whether the project used ESLint; an
// error, with nothing written, when package.json cannot be read.
func SplitNode(req Request) (Result, error) {
	var res Result
	dir := filepath.Join(req.Root, filepath.FromSlash(req.Dir))
	raw, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return res, err
	}
	project, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return res, fmt.Errorf("%s/package.json: %w", req.Dir, err)
	}
	rootPath := filepath.Join(req.Root, "package.json")
	rootRaw, err := os.ReadFile(rootPath)
	if err != nil {
		rootRaw = []byte("{}")
	}
	root, err := jsondoc.ParseDocument(rootRaw)
	if err != nil {
		return res, fmt.Errorf("package.json at the root: %w", err)
	}

	hoisted := takeDevDependencies(project, req.Hoist)
	dropped := takeDevDependencies(project, req.Drop)
	if !root.Has("name") {
		root.SetString("name", req.Workspace+"-workspace")
	}
	if !root.Has("private") {
		_ = root.SetJSON("private", true)
	}
	if !root.Has("workspaces") {
		_ = root.SetJSON("workspaces", req.Patterns)
	}
	if len(hoisted) > 0 {
		dev := root.Object("devDependencies", true)
		for _, name := range slices.Sorted(mapKeys(hoisted)) {
			if _, has := dev.String(name); !has {
				dev.SetString(name, hoisted[name])
			}
		}
	}

	if err := os.WriteFile(filepath.Join(dir, "package.json"), project.Bytes(), 0o644); err != nil {
		return res, err
	}
	if err := os.WriteFile(rootPath, root.Bytes(), 0o644); err != nil {
		return res, err
	}
	if len(hoisted) > 0 {
		res.Notes = append(res.Notes, fmt.Sprintf("moved %d devDependencies to the root package.json: %s.", len(hoisted), strings.Join(slices.Sorted(mapKeys(hoisted)), ", ")))
	}
	if len(dropped) > 0 {
		res.Notes = append(res.Notes, "removed from the project, because the shared lint config brings its own: "+strings.Join(slices.Sorted(mapKeys(dropped)), ", ")+".")
		res.ESLint = true
	}

	note, err := shareTSConfig(req, dir)
	if err != nil {
		return res, err
	}
	if note != "" {
		res.Notes = append(res.Notes, note)
	}
	build, err := moveBuildIntoDist(req)
	if err != nil {
		return res, err
	}
	if build != "" {
		res.Notes = append(res.Notes, build)
	}
	legacy, err := retireLegacyESLint(req.Root)
	if err != nil {
		return res, err
	}
	for _, name := range legacy {
		res.Notes = append(res.Notes, name+" is kept as "+name+".pre-smith: the workspace lints with the shared flat config, which ESLint reads instead.")
		res.ESLint = true
	}
	if flat, _ := filepath.Glob(filepath.Join(req.Root, "eslint.config.*")); len(flat) > 0 {
		res.ESLint = true
	}
	if jest, _ := filepath.Glob(filepath.Join(dir, "jest.config.*")); len(jest) > 0 {
		res.Notes = append(res.Notes, filepath.Base(jest[0])+" stays with the project: its paths are relative to the project folder. The jest packages are in the workspace.")
	}
	return res, nil
}

// takeDevDependencies removes the named devDependencies from the manifest
// and returns each with its version. An emptied devDependencies is removed.
func takeDevDependencies(doc *jsondoc.Document, names []string) map[string]string {
	taken := map[string]string{}
	dev := doc.Object("devDependencies", false)
	if dev == nil {
		return taken
	}
	for _, name := range names {
		if version, ok := dev.String(name); ok {
			taken[name] = version
			dev.Delete(name)
		}
	}
	if dev.Len() == 0 {
		doc.Delete("devDependencies")
	}
	return taken
}

func mapKeys[V any](m map[string]V) func(yield func(string) bool) {
	return func(yield func(string) bool) {
		for k := range m {
			if !yield(k) {
				return
			}
		}
	}
}

// retireLegacyESLint renames the legacy ESLint configuration files at the
// root (.eslintrc, .eslintrc.json, ...) to <name>.pre-smith and returns their
// names.
func retireLegacyESLint(root string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(root, ".eslintrc*"))
	if err != nil {
		return nil, err
	}
	var names []string
	for _, m := range matches {
		name := filepath.Base(m)
		if strings.HasSuffix(name, ".pre-smith") {
			continue
		}
		if err := os.Rename(m, m+".pre-smith"); err != nil {
			return names, err
		}
		names = append(names, name)
	}
	slices.Sort(names)
	return names, nil
}
