package tooling

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Installer runs npm install at the workspace root; tests replace it.
var Installer = func(workspaceRoot string) error {
	return toolrun.Quiet(workspaceRoot, []string{"npm", "install", "--no-audit", "--no-fund"})
}

// Apply adds s to the workspace and to the project params names, and installs
// when the root devDependencies changed. It returns what it wrote,
// workspace-relative.
func Apply(workspaceRoot string, params protocol.GenerateParams, s Setup) (protocol.GenerateResult, error) {
	written, err := writeAbsent(workspaceRoot, "", s.RootFiles)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	rootChanged, err := edit(filepath.Join(workspaceRoot, "package.json"), func(doc *jsondoc.Document) (bool, error) {
		changed := addStrings(doc.Object("devDependencies", true), s.DevDependencies)
		added, err := addFields(doc, s.RootFields)
		return changed || added, err
	})
	if err != nil {
		return protocol.GenerateResult{Files: written}, fmt.Errorf("root package.json: %w", err)
	}
	if rootChanged {
		written = append(written, "package.json")
	}
	files, err := writeAbsent(workspaceRoot, params.Directory, s.Files)
	written = append(written, files...)
	if err != nil {
		return protocol.GenerateResult{Files: written}, err
	}
	manifest := path.Join(params.Directory, "package.json")
	projectChanged, err := edit(filepath.Join(workspaceRoot, filepath.FromSlash(manifest)), func(doc *jsondoc.Document) (bool, error) {
		changed := addStrings(doc.Object("scripts", true), s.Scripts)
		added, err := addFields(doc, s.Fields)
		return changed || added, err
	})
	if err != nil {
		return protocol.GenerateResult{Files: written}, fmt.Errorf("%s: %w", manifest, err)
	}
	if projectChanged {
		written = append(written, manifest)
	}
	sort.Strings(written)
	if rootChanged {
		if err := Installer(workspaceRoot); err != nil {
			return protocol.GenerateResult{Files: written}, fmt.Errorf("installing dependencies: %w", err)
		}
	}
	return protocol.GenerateResult{Files: written}, nil
}

// edit rewrites a package.json when change reports a change. The file must
// exist: the node language plugin writes it.
func edit(file string, change func(*jsondoc.Document) (bool, error)) (bool, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return false, err
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return false, err
	}
	changed, err := change(doc)
	if err != nil || !changed {
		return false, err
	}
	return true, os.WriteFile(file, doc.Bytes(), 0o644)
}

func addStrings(doc *jsondoc.Document, values map[string]string) bool {
	changed := false
	for _, key := range sortedKeys(values) {
		if !doc.Has(key) {
			doc.SetString(key, values[key])
			changed = true
		}
	}
	return changed
}

// writeAbsent writes each file under root/dir that does not exist yet.
func writeAbsent(root, dir string, files map[string]string) ([]string, error) {
	var written []string
	for _, name := range sortedKeys(files) {
		rel := path.Join(dir, name)
		p := filepath.Join(root, filepath.FromSlash(rel))
		if _, err := os.Stat(p); err == nil {
			continue
		} else if !errors.Is(err, fs.ErrNotExist) {
			return written, err
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return written, err
		}
		if err := os.WriteFile(p, []byte(files[name]), 0o644); err != nil {
			return written, err
		}
		written = append(written, rel)
	}
	return written, nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// addFields adds the fields doc lacks. A field doc has stays as it is, except
// that an object gains the entries it lacks: two toolchains can each add their
// own npm "overrides" without the second finding the field taken.
func addFields(doc *jsondoc.Document, fields map[string]any) (bool, error) {
	changed := false
	for _, key := range sortedKeys(fields) {
		if !doc.Has(key) {
			if err := doc.SetJSON(key, fields[key]); err != nil {
				return false, err
			}
			changed = true
			continue
		}
		wanted, isObject := fields[key].(map[string]any)
		existing := doc.Object(key, false)
		if !isObject || existing == nil {
			continue
		}
		added, err := addFields(existing, wanted)
		if err != nil {
			return false, err
		}
		changed = changed || added
	}
	return changed, nil
}
