package manifestfixes

import (
	"path"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// cleanDir is a folder as a slash path without a leading "./".
func cleanDir(dir string) string {
	return strings.TrimPrefix(path.Clean(strings.ReplaceAll(dir, "\\", "/")), "./")
}

// under reports whether a path is the folder or inside it.
func under(value, dir string) bool {
	value = cleanDir(value)
	return dir != "" && dir != "." && (value == dir || strings.HasPrefix(value, dir+"/"))
}

// followPath moves a path that is the old folder, or inside it, to the new one.
func followPath(value, old, newDir string) string {
	if !under(value, old) {
		return value
	}
	rest := strings.TrimPrefix(strings.TrimPrefix(cleanDir(value), old), "/")
	if rest == "" {
		return newDir
	}
	return newDir + "/" + rest
}

// rewriteFlags moves the folders and files that the flags of a build script
// name (--outDir, --outdir, --outfile, --declarationDir) from the old folder to
// the new one, in place. Every other character of the script stays.
func rewriteFlags(script, old, newDir string) string {
	return pathFlag.ReplaceAllStringFunc(script, func(m string) string {
		parts := pathFlag.FindStringSubmatch(m)
		return parts[1] + parts[2] + followPath(parts[3], old, newDir)
	})
}

// rewriteScriptPaths moves the words of a script that are a path into the old
// folder (dist/main.js, ./dist, 'dist/index.js') to the new one. A bare word
// that only looks like the folder ("next build") is not a path and stays.
func rewriteScriptPaths(script, old, newDir string) string {
	if old == "" || old == "." {
		return script
	}
	word := regexp.MustCompile(`(^|[\s'"=(])(\./` + regexp.QuoteMeta(old) + `(?:/|[\s'")]|$)|` + regexp.QuoteMeta(old) + `/)`)
	return word.ReplaceAllStringFunc(script, func(m string) string {
		lead := m[:len(m)-len(strings.TrimLeft(m, " \t'\"=("))]
		body := strings.TrimPrefix(m[len(lead):], "./")
		rest := strings.TrimPrefix(body, old)
		return lead + newDir + rest
	})
}

// rewriteConfigPaths moves the declarationDir, tsBuildInfoFile and outFile of
// a tsconfig from the old folder to the new one.
func rewriteConfigPaths(config, old, newDir string) string {
	return configPath.ReplaceAllStringFunc(config, func(m string) string {
		parts := configPath.FindStringSubmatch(m)
		return parts[1] + `"` + followPath(parts[2], old, newDir) + `"`
	})
}

// listsOnly reports whether the manifest has a files list and every entry is
// the old build folder or inside it.
func listsOnly(doc *jsondoc.Document, old string) bool {
	entries := fileEntries(doc)
	for _, e := range entries {
		if !under(e, old) {
			return false
		}
	}
	return len(entries) > 0
}

// fileEntries are the strings of the manifest's files list.
func fileEntries(doc *jsondoc.Document) []string {
	var out []string
	doc.RewriteStrings([]string{"files"}, func(v string) string { out = append(out, v); return v })
	return out
}
