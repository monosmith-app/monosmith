package recipecatalog

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
)

// Suffix ends the name of a recipe file.
const Suffix = ".recipe.toml"

// Listed is a recipe file found in a workspace.
type Listed struct {
	// Path is the file, relative to the workspace root, with forward slashes.
	Path    string `json:"path"`
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	Extends string `json:"extends,omitempty"`
	// Error says why the file could not be read as a recipe.
	Error string `json:"error,omitempty"`
}

// skippedDirs hold other tools' files, not the workspace's recipes.
var skippedDirs = map[string]bool{"node_modules": true, "dist": true, "vendor": true, "target": true}

// ListRecipes finds the recipe files (*.recipe.toml) under root, in path
// order, and says what each is. A file that is not a valid recipe is listed
// with its error.
func ListRecipes(root string) ([]Listed, error) {
	listed := []Listed{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if path != root && (strings.HasPrefix(name, ".") || skippedDirs[name]) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(name, Suffix) {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		entry := Listed{Path: filepath.ToSlash(rel)}
		data, err := os.ReadFile(path)
		if err == nil {
			var doc recipe.Document
			if doc, err = recipe.ParseRecipe(data, recipe.Source{Ref: entry.Path}); err == nil {
				entry.Name, entry.Version, entry.Extends = doc.Recipe.Name, doc.Recipe.Version, doc.Recipe.Extends
			}
		}
		if err != nil {
			entry.Error = err.Error()
		}
		listed = append(listed, entry)
		return nil
	})
	sort.Slice(listed, func(i, j int) bool { return listed[i].Path < listed[j].Path })
	return listed, err
}
