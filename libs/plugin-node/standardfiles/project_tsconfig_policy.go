package standardfiles

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/tsconfigchain"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// The tsconfig files of a project smith writes (the typescript stack).
const (
	TsconfigJSON  = "tsconfig.json"
	TsconfigBuild = "tsconfig.build.json"
	// tsconfigLib marks the Nx layout, whose tsconfig.json only references
	// tsconfig.lib.json: the standard leaves it alone.
	tsconfigLib = "tsconfig.lib.json"
)

// ProjectTsconfig is the tsconfig.json of a project, as the typescript stack
// writes it: the workspace's base, its sources in src, Node's types.
func ProjectTsconfig(dir string) string {
	return `{
  "extends": "` + tsconfigchain.Extends(dir, tsconfigchain.Base) + `",
  "compilerOptions": {
    "rootDir": "src",
    "rewriteRelativeImportExtensions": true,
    "types": ["node"]
  },
  "include": ["src"]
}
`
}

// BuildTsconfig is the tsconfig.build.json of a project: its tsconfig.json
// without the tests.
func BuildTsconfig(testSuffix string) string {
	return `{
  "extends": "./tsconfig.json",
  "exclude": ["src/**/*.` + testSuffix + `.ts"]
}
`
}

// emitOptions are compiler options that only say where and what the build
// writes. The standard leaves them to the build command and the base, so a
// tsconfig.json that differs from it only by them compiles the code the same.
var emitOptions = []string{"declaration", "declarationMap", "sourceMap", "tsBuildInfoFile", "incremental"}

// ReplaceableTsconfig says whether a project's tsconfig.json can become the
// standard without changing how its code compiles: it extends the workspace's
// base, compiles src, and sets no compiler option but those the standard sets
// the same way, emit options, and outDir when the build passes --outDir.
// A browser project (jsx, the DOM), a project with paths, references or an
// option of its own (strict, target, ...) keeps its file.
//
// Parameters: dir is the project folder, workspace-relative; raw is its
// tsconfig.json; build is its build script ("" for none).
func ReplaceableTsconfig(dir string, raw []byte, build string) bool {
	clean := jsondoc.StripComments(raw)
	var doc map[string]json.RawMessage
	if json.Unmarshal(clean, &doc) != nil {
		return false
	}
	for key, value := range doc {
		switch key {
		case "$schema", "compilerOptions":
		case "extends":
			var s string
			if json.Unmarshal(value, &s) != nil || s != tsconfigchain.Extends(dir, tsconfigchain.Base) {
				return false
			}
		case "include":
			if !sourcesOnly(value) {
				return false
			}
		case "exclude":
			if !buildFoldersOnly(value) {
				return false
			}
		default:
			return false
		}
	}
	if _, ok := doc["extends"]; !ok {
		return false
	}
	var options map[string]json.RawMessage
	if raw, ok := doc["compilerOptions"]; ok && json.Unmarshal(raw, &options) != nil {
		return false
	}
	for key, value := range options {
		var s string
		var b bool
		var list []string
		switch {
		case key == "rootDir":
			if json.Unmarshal(value, &s) != nil || strings.TrimPrefix(strings.TrimSuffix(s, "/"), "./") != "src" {
				return false
			}
		case key == "rewriteRelativeImportExtensions":
			if json.Unmarshal(value, &b) != nil || !b {
				return false
			}
		case key == "types":
			if json.Unmarshal(value, &list) != nil || !slices.Equal(list, []string{"node"}) {
				return false
			}
		case key == "outDir":
			if build != "" && !strings.Contains(build, "--outDir") {
				return false
			}
		case slices.Contains(emitOptions, key):
		default:
			return false
		}
	}
	return true
}

// ReplaceableBuildTsconfig says whether a project's tsconfig.build.json can
// become the standard: it extends ./tsconfig.json and only leaves tests out,
// with at most emit options of its own.
func ReplaceableBuildTsconfig(raw []byte) bool {
	var doc map[string]json.RawMessage
	if json.Unmarshal(jsondoc.StripComments(raw), &doc) != nil {
		return false
	}
	for key, value := range doc {
		switch key {
		case "$schema", "exclude":
		case "extends":
			var s string
			if json.Unmarshal(value, &s) != nil || strings.TrimPrefix(s, "./") != TsconfigJSON {
				return false
			}
		case "include":
			if !sourcesOnly(value) {
				return false
			}
		case "compilerOptions":
			var options map[string]json.RawMessage
			if json.Unmarshal(value, &options) != nil {
				return false
			}
			for k := range options {
				if !slices.Contains(emitOptions, k) && k != "outDir" {
					return false
				}
			}
		default:
			return false
		}
	}
	return true
}

// nxLayout says whether the project keeps its build options in
// tsconfig.lib.json (the Nx layout), which the standard leaves alone.
func nxLayout(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, tsconfigLib))
	return err == nil
}

// sourcesOnly says whether an "include" names only the src folder.
func sourcesOnly(value json.RawMessage) bool {
	var list []string
	if json.Unmarshal(value, &list) != nil || len(list) == 0 {
		return false
	}
	for _, p := range list {
		switch strings.TrimPrefix(p, "./") {
		case "src", "src/", "src/**/*", "src/**/*.ts":
		default:
			return false
		}
	}
	return true
}

// buildFoldersOnly says whether an "exclude" names only folders that hold no
// sources (node_modules, a build folder), which "include": ["src"] leaves out
// anyway.
func buildFoldersOnly(value json.RawMessage) bool {
	var list []string
	if json.Unmarshal(value, &list) != nil {
		return false
	}
	for _, p := range list {
		switch strings.Trim(strings.TrimPrefix(p, "./"), "/") {
		case "node_modules", "dist", "build", "out", "lib":
		default:
			return false
		}
	}
	return true
}
