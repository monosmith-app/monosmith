package discovery

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Manifest is the name pattern of a recipe file.
const Manifest = "*.recipe.toml"

// Suffix ends the name of a recipe file.
const Suffix = ".recipe.toml"

// File is what a project needs of a recipe file: its name, its version and
// the recipe it extends.
type File struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
	Extends string `toml:"extends"`
}

// Read reads the top of a recipe file. Other keys are left to the recipe
// loader, which checks them.
func Read(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}
	var f File
	if err := toml.Unmarshal(data, &f); err != nil {
		return File{}, err
	}
	return f, nil
}

// OwnFolder reports whether a folder is a recipe's own: it holds exactly one
// recipe file, and besides it only documentation (README, CHANGELOG, LICENSE)
// and hidden files. A folder with anything else is another project's, or
// holds several recipes that would be released together, and is not a recipe
// project. It returns the recipe file's name.
func OwnFolder(dir string) (file string, ok bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", false
	}
	for _, e := range entries {
		name := e.Name()
		switch {
		case strings.HasPrefix(name, "."):
			continue
		case e.IsDir():
			return "", false
		case strings.HasSuffix(name, Suffix):
			if file != "" {
				return "", false
			}
			file = name
		case isDoc(name):
		default:
			return "", false
		}
	}
	return file, file != ""
}

func isDoc(name string) bool {
	stem := strings.ToUpper(strings.TrimSuffix(name, filepath.Ext(name)))
	switch stem {
	case "README", "CHANGELOG", "LICENSE", "LICENCE", "NOTICE":
		return true
	}
	return false
}
