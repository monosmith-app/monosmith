package tsconfigchain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// PathAliases are the compilerOptions "paths" keys ("@shared/*", "~/*") of the
// tsconfig files in dir (tsconfig.json, tsconfig.app.json and the others named
// tsconfig*.json) and of every file they extend by a relative path. An import
// an alias matches is the project's own code, not a package.
//
// Parameters: dir is the project folder.
// Returns: the keys, in no set order; none when no file declares paths or a
// file cannot be read.
func PathAliases(dir string) []string {
	files, _ := filepath.Glob(filepath.Join(dir, "tsconfig*.json"))
	seen := map[string]bool{}
	var out []string
	for _, f := range files {
		collectPaths(f, 0, seen, &out)
	}
	return out
}

func collectPaths(file string, depth int, seen map[string]bool, out *[]string) {
	file = filepath.Clean(file)
	if depth > maxDepth || seen[file] {
		return
	}
	seen[file] = true
	raw, err := os.ReadFile(file)
	if err != nil {
		return
	}
	var doc struct {
		Extends         json.RawMessage `json:"extends"`
		CompilerOptions struct {
			Paths map[string]json.RawMessage `json:"paths"`
		} `json:"compilerOptions"`
	}
	if json.Unmarshal(jsondoc.StripComments(raw), &doc) != nil {
		return
	}
	for key := range doc.CompilerOptions.Paths {
		*out = append(*out, key)
	}
	for _, target := range extendsOf(doc.Extends) {
		if !strings.HasPrefix(target, ".") {
			continue
		}
		if !strings.HasSuffix(target, ".json") {
			target += ".json"
		}
		collectPaths(filepath.Join(filepath.Dir(file), filepath.FromSlash(target)), depth+1, seen, out)
	}
}

// Aliased is true when the package name an import was read as ("@shared/contracts"
// for "@shared/contracts/x") is one of the aliases: "@shared/*" matches every
// name under "@shared/", and a key with no "*" matches that name exactly.
//
// Parameters: name is the package name of an import; aliases are PathAliases.
// Returns: whether an alias covers it.
func Aliased(name string, aliases []string) bool {
	for _, a := range aliases {
		if prefix, wild := strings.CutSuffix(a, "*"); wild {
			if prefix == "" || strings.HasPrefix(name+"/", prefix) || strings.HasPrefix(name, prefix) {
				return true
			}
		} else if name == a || strings.HasPrefix(a, name+"/") {
			return true
		}
	}
	return false
}
