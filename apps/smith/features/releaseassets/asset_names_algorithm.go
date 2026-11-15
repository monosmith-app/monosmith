package releaseassets

import (
	"fmt"
	"path/filepath"
	"strings"
)

// compoundExtensions are extensions with a dot in them, for {ext}.
var compoundExtensions = []string{".app.tar.gz", ".tar.gz", ".tar.xz", ".tar.bz2"}

// assetNames names each file for upload: pattern with {product}, {version},
// {platform} and {ext} filled in. Files that get the same name (a .exe from two
// bundle kinds, say) keep their folder's name before the extension, so every
// runner's names are unique. An empty pattern keeps the files' own names.
func assetNames(files []string, pattern, product, version, platform string) (map[string]string, error) {
	names := make(map[string]string, len(files))
	if pattern == "" {
		for _, f := range files {
			names[f] = filepath.Base(f)
		}
		return names, nil
	}
	byName := map[string][]string{}
	for _, f := range files {
		name := strings.NewReplacer("{product}", product, "{version}", version, "{platform}", platform, "{ext}", extensionOf(f)).Replace(pattern)
		byName[name] = append(byName[name], f)
	}
	for name, group := range byName {
		if len(group) == 1 {
			names[group[0]] = name
			continue
		}
		for _, f := range group {
			folder := filepath.Base(filepath.Dir(f))
			if ext := extensionOf(f); ext != "" && strings.HasSuffix(name, "."+ext) {
				names[f] = strings.TrimSuffix(name, "."+ext) + "-" + folder + "." + ext
			} else {
				names[f] = name + "-" + folder
			}
		}
	}
	seen := map[string]string{}
	for f, name := range names {
		if other, dup := seen[name]; dup {
			return nil, fmt.Errorf("%s and %s would both be uploaded as %s; change asset_name or assets", other, f, name)
		}
		seen[name] = f
	}
	return names, nil
}

// extensionOf is the file's extension without the dot, counting .tar.gz as one.
func extensionOf(file string) string {
	base := filepath.Base(file)
	for _, c := range compoundExtensions {
		if strings.HasSuffix(base, c) {
			return c[1:]
		}
	}
	return strings.TrimPrefix(filepath.Ext(base), ".")
}
