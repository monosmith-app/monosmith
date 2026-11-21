package configedit

import (
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
)

// manifestNames are the files that make a folder a project of a built-in
// plugin, in the order they are looked for; manifestGlobs the ones whose name
// varies (a .NET project file, a gemspec).
var (
	manifestNames = []string{
		"package.json", "go.mod", "pyproject.toml", "Cargo.toml", "pom.xml", "build.gradle.kts", "build.gradle",
		"pubspec.yaml", "composer.json", "Package.swift", "CMakeLists.txt", "main.bicep", "app.json",
	}
	manifestGlobs = []string{"*.csproj", "*.fsproj", "*.vbproj", "*.gemspec", "*.rockspec", "*.podspec", "*.recipe.toml"}
)

// ReadProjectSettings reads what the editors show for one project of the
// workspace at root: the values its manifest owns (read-only, with the file
// that holds them) and the settings of [projects.<name>] with their values.
//
// Parameters: root is the workspace root; p is what the project's plugin
// reports about it.
// Returns: the view; an error when smith.toml cannot be read or loaded.
func ReadProjectSettings(root string, p ProjectFacts) (ProjectView, error) {
	cfg, err := workspace.Load(root)
	if err != nil {
		return ProjectView{}, err
	}
	raw, err := os.ReadFile(filepath.Join(root, workspace.FileName))
	if err != nil {
		return ProjectView{}, err
	}
	v := ProjectView{Name: p.Name, Declared: cfg.Projects[p.Name].Root != ""}
	v.Manifest = workspace.FileName
	if !v.Declared {
		v.Manifest = findManifest(root, p.Root)
	}
	fact := func(key, value, explanation string) {
		if value != "" {
			v.Facts = append(v.Facts, ProjectFact{Key: key, Value: value, Explanation: explanation, Source: v.Manifest})
		}
	}
	if v.Declared {
		fact("name", p.Name, "The project's name: the name of its [projects."+TableKey(p.Name)+"] table in smith.toml. Rename the table to rename it.")
		fact("version", p.Version, "The version the project's plugin reports.")
	} else {
		fact("name", p.Name, "The project's name, as its manifest gives it. Change it in the manifest.")
		fact("version", p.Version, "The version its manifest records. A release changes it.")
		fact("language", p.Language, "The plugin that found the project by its manifest.")
		fact("kind", p.Kind, "What the project is: app, lib or package, from its manifest. Change kind (smith kind) changes it.")
		fact("root", p.Root, "The folder of the project: where its manifest is.")
	}
	fields, err := ProjectFields(p.Name)
	if err != nil {
		return ProjectView{}, err
	}
	for _, f := range fields {
		if f.DeclaredOnly && !v.Declared {
			continue // the manifest and the plugin give them (see the facts)
		}
		value, set := GetKey(string(raw), f.Table(), f.Key())
		v.Settings = append(v.Settings, ProjectSetting{Field: f, Value: value, Set: set})
	}
	return v, nil
}

// findManifest is the workspace-relative path of the manifest in the
// project's folder, or the folder itself (with a slash) when no file smith
// knows is there.
func findManifest(root, dir string) string {
	folder := filepath.Join(root, filepath.FromSlash(dir))
	for _, name := range manifestNames {
		if _, err := os.Stat(filepath.Join(folder, name)); err == nil {
			return path.Join(dir, name)
		}
	}
	for _, glob := range manifestGlobs {
		if found, _ := filepath.Glob(filepath.Join(folder, glob)); len(found) > 0 {
			return path.Join(dir, filepath.Base(found[0]))
		}
	}
	return path.Clean(dir) + "/"
}
