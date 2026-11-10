package doctorview

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
)

// FileOf is the file a finding is about, for the button that asks smith why
// it is here: the first file its fix writes, else the first word of its
// problem or detail that reads as a path ("packages/web/tsconfig.json",
// "smith.toml"). It is "" for a finding about no file.
func FileOf(f doctor.Finding) string {
	if f.Fix != nil {
		for _, e := range f.Fix.Files {
			if e.Path != "" {
				return e.Path
			}
		}
	}
	for _, text := range []string{f.Problem, f.Detail} {
		for _, field := range strings.Fields(text) {
			// A sentence's full stop goes; the dot that starts ".gitignore" stays.
			p := strings.TrimRight(strings.Trim(field, "`'\"()[],;:"), ".,;:`'\")")
			if looksLikePath(p) {
				return p
			}
		}
	}
	return ""
}

// looksLikePath is true for a word that names a file: its last part has a name
// and an extension of letters ("tsconfig.json", ".gitignore"), it is not a URL,
// a flag or a version ("2.15.0"), and it is not an abbreviation ("e.g").
func looksLikePath(p string) bool {
	if p == "" || strings.HasPrefix(p, "-") || strings.Contains(p, "://") || strings.ContainsAny(p, "<>{}=*") {
		return false
	}
	base := p[strings.LastIndexAny(p, `/\`)+1:]
	dot := strings.LastIndex(base, ".")
	if dot < 0 || dot == len(base)-1 {
		return false
	}
	name, ext := base[:dot], base[dot+1:]
	if name != "" && len(name) < 2 {
		return false
	}
	for _, r := range ext {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	if strings.ContainsAny(p, `/\`) {
		return true
	}
	// A bare name is a file only with an extension files have, and a product
	// name such as "Node.js" is not one.
	product := ext == "js" && name != "" && name[0] >= 'A' && name[0] <= 'Z'
	return fileExtensions[ext] && !product
}

// fileExtensions are the extensions that make a bare word a file name.
var fileExtensions = map[string]bool{
	"toml": true, "json": true, "jsonc": true, "yaml": true, "yml": true, "mod": true, "sum": true, "lock": true,
	"md": true, "txt": true, "cfg": true, "ini": true, "xml": true, "gradle": true, "kts": true, "csproj": true,
	"props": true, "ts": true, "tsx": true, "js": true, "mjs": true, "cjs": true, "go": true, "py": true,
	"rs": true, "dart": true, "gitignore": true, "npmrc": true, "env": true,
}
