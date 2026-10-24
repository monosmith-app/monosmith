package buildoutputs

import (
	"path"
	"regexp"
	"slices"
	"strings"
)

// dirFlags take a directory; fileFlags a file, whose directory is written.
// -o and --output take either, told apart by an extension.
var (
	dirFlags    = map[string]bool{"--outdir": true, "--out-dir": true, "--outDir": true, "--declarationDir": true}
	fileFlags   = map[string]bool{"--outfile": true, "--out-file": true, "--outFile": true}
	eitherFlags = map[string]bool{"-o": true, "--output": true}
	// inlineOutdir is the outdir option of the node -e one-liner convert writes.
	inlineOutdir = regexp.MustCompile(`\boutdir\s*:\s*['"]([^'"]+)['"]`)
	shellSplit   = strings.NewReplacer("&&", "\n", "||", "\n", ";", "\n", "|", "\n")
)

// WrittenDirs returns the directories, relative to the project, a build
// script writes its output to, cleaned and without duplicates. The project
// root itself is left out, since nothing can be declared there, and so is
// any value a shell would expand.
func WrittenDirs(script string) []string {
	var out []string
	add := func(dir string, isFile bool) {
		dir = strings.Trim(dir, `'"`)
		if dir == "" || strings.ContainsAny(dir, "$*?{}()<>`\\!~ ") || path.IsAbs(dir) || strings.HasPrefix(dir, "-") {
			return
		}
		dir = path.Clean(dir)
		if isFile {
			dir = path.Dir(dir)
		}
		if dir == "." || dir == ".." || slices.Contains(out, dir) {
			return
		}
		out = append(out, dir)
	}
	for _, m := range inlineOutdir.FindAllStringSubmatch(script, -1) {
		add(m[1], false)
	}
	for _, segment := range strings.Split(shellSplit.Replace(script), "\n") {
		words := strings.Fields(segment)
		for i, w := range words {
			flag, value, inline := strings.Cut(w, "=")
			if !dirFlags[flag] && !fileFlags[flag] && !eitherFlags[flag] {
				continue
			}
			if !inline {
				if i+1 >= len(words) {
					continue
				}
				value = words[i+1]
			}
			switch {
			case dirFlags[flag]:
				add(value, false)
			case fileFlags[flag]:
				add(value, true)
			default:
				add(value, path.Ext(strings.Trim(value, `'"`)) != "")
			}
		}
	}
	return out
}

// Collapse drops every directory that lies inside another one in the list.
func Collapse(dirs []string) []string {
	var out []string
	for _, d := range dirs {
		inside := false
		for _, other := range dirs {
			if other != d && strings.HasPrefix(d, other+"/") {
				inside = true
			}
		}
		if !inside && !slices.Contains(out, d) {
			out = append(out, d)
		}
	}
	return out
}
