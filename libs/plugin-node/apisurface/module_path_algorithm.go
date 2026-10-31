package apisurface

import (
	"path"
	"strings"
)

// moduleExtensions are the file kinds tried for a module, source first.
var moduleExtensions = []string{".ts", ".tsx", ".d.ts", ".mts", ".cts", ".js", ".mjs", ".cjs", ".jsx"}

// buildDirs are the directories a build writes to, and subdirectories of
// them that hold one module format or the declarations.
var (
	buildDirs = []string{"dist", "lib", "build", "out"}
	formatDir = []string{"esm", "cjs", "es", "umd", "types", "typings", "browser", "node", "src"}
)

// resolveModule finds the file a relative specifier ("./x", "../y/index.js")
// in a file stands for. TypeScript sources name the built file ("./x.js"), so
// a .js specifier also matches x.ts.
func resolveModule(src Source, fromFile, spec string) (string, bool) {
	if !strings.HasPrefix(spec, ".") {
		return "", false
	}
	base := path.Join(path.Dir(fromFile), spec)
	var stems []string
	if ext := path.Ext(base); ext == ".js" || ext == ".mjs" || ext == ".cjs" || ext == ".jsx" {
		stems = append(stems, strings.TrimSuffix(base, ext))
	}
	stems = append(stems, base)
	for _, stem := range stems {
		for _, ext := range moduleExtensions {
			if _, ok := src.Read(stem + ext); ok {
				return stem + ext, true
			}
		}
	}
	if _, ok := src.Read(base); ok && path.Ext(base) != "" {
		return base, true
	}
	for _, ext := range moduleExtensions {
		if _, ok := src.Read(path.Join(base, "index") + ext); ok {
			return path.Join(base, "index") + ext, true
		}
	}
	return "", false
}

// sourceVariants lists the files a built path may come from, when the built
// file is not in the repository: dist/esm/index.js is src/index.ts, or
// index.ts next to the package.json. The dir is the package directory, which
// the paths are relative to.
func sourceVariants(file string) []string {
	rest := file
	parts := strings.Split(file, "/")
	if len(parts) > 1 && containsString(buildDirs, parts[0]) {
		rest = strings.Join(parts[1:], "/")
		if more := strings.Split(rest, "/"); len(more) > 1 && containsString(formatDir, more[0]) {
			rest = strings.Join(more[1:], "/")
		}
	}
	stem := strings.TrimSuffix(strings.TrimSuffix(rest, ".d.ts"), path.Ext(strings.TrimSuffix(rest, ".d.ts")))
	var out []string
	for _, dir := range []string{"src/", ""} {
		for _, ext := range moduleExtensions {
			if candidate := dir + stem + ext; candidate != file {
				out = append(out, candidate)
			}
		}
	}
	return out
}
