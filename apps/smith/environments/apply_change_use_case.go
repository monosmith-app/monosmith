package environments

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/scaffold"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// Apply writes a change: the new env files (never over one that exists), the
// scripts of the app's package.json and the list in smith.toml. It stops at the
// first error; what was written stays, and smith.toml is written last, so a
// failure before it leaves the recorded environments as they were.
//
// Parameters: root is the workspace folder; c the change from PlanChange.
// Returns: the workspace-relative paths written; an error when a file cannot be
// read or written, or package.json is not a JSON object.
func Apply(root string, c Change) ([]string, error) {
	var written []string
	appDir := filepath.Join(root, filepath.FromSlash(c.Root))
	for _, name := range slices.Sorted(mapKeys(c.Created)) {
		path := filepath.Join(appDir, name)
		if _, err := os.Stat(path); err == nil {
			continue
		} else if !errors.Is(err, fs.ErrNotExist) {
			return written, err
		}
		if err := os.WriteFile(path, []byte(c.Created[name]), 0o644); err != nil {
			return written, err
		}
		written = append(written, c.Root+"/"+name)
	}
	manifest := filepath.Join(appDir, "package.json")
	raw, err := os.ReadFile(manifest)
	if err != nil {
		return written, err
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return written, fmt.Errorf("%s/package.json: %w", c.Root, err)
	}
	scripts := doc.Object("scripts", true)
	for _, name := range slices.Sorted(mapKeys(c.ScriptsAdded)) {
		scripts.SetString(name, c.ScriptsAdded[name])
	}
	for _, name := range c.ScriptsRemoved {
		scripts.Delete(name)
	}
	if err := os.WriteFile(manifest, doc.Bytes(), 0o644); err != nil {
		return written, err
	}
	written = append(written, c.Root+"/package.json")
	tomlPath := filepath.Join(root, workspace.FileName)
	toml, err := os.ReadFile(tomlPath)
	if err != nil {
		return written, err
	}
	text := scaffold.EnvironmentsToml(string(toml), c.Project, c.After)
	if _, err := workspace.Parse([]byte(text)); err != nil {
		return written, err
	}
	if err := os.WriteFile(tomlPath, []byte(text), 0o644); err != nil {
		return written, err
	}
	return append(written, workspace.FileName), nil
}
