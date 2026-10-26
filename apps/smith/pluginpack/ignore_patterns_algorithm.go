package pluginpack

import (
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// IgnoreFile is the file at the root of a plugin directory that lists what
// "smith plugin pack" leaves out.
const IgnoreFile = ".smithpluginignore"

// ignorePattern is one line of the ignore file.
type ignorePattern struct {
	glob    string
	dirOnly bool
	// anywhere is a pattern without a slash: it matches the name at any depth.
	anywhere bool
}

// parseIgnore reads the ignore file's lines. A blank line and a line starting
// with # are skipped. "name" matches that name at any depth, "dir/" only a
// directory, "a/b" and "/a" a path from the plugin's root; * and ** are globs.
// There is no negation: what is listed is left out.
func parseIgnore(text string) []ignorePattern {
	var out []ignorePattern
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p := ignorePattern{}
		if strings.HasSuffix(line, "/") {
			p.dirOnly = true
			line = strings.TrimSuffix(line, "/")
		}
		p.anywhere = !strings.Contains(line, "/")
		p.glob = strings.TrimPrefix(line, "/")
		if p.glob != "" {
			out = append(out, p)
		}
	}
	return out
}

// ignored says whether the entry at rel (slash-separated, from the plugin's
// root) is left out.
func ignored(patterns []ignorePattern, rel string, isDir bool) bool {
	base := rel[strings.LastIndex(rel, "/")+1:]
	for _, p := range patterns {
		if p.dirOnly && !isDir {
			continue
		}
		target := rel
		if p.anywhere {
			target = base
		}
		if ok, _ := doublestar.Match(p.glob, target); ok {
			return true
		}
	}
	return false
}
