package discovery

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	sourceExtensions = map[string]bool{".ts": true, ".tsx": true, ".mts": true, ".cts": true, ".js": true, ".jsx": true, ".mjs": true, ".cjs": true}
	skippedDirs      = map[string]bool{"node_modules": true, "dist": true, "coverage": true, "build": true}
	specifierPattern = regexp.MustCompile(`(?:\bfrom\s*|\bimport\s*\(?\s*|\brequire\s*\(\s*)['"]([^'"\n]+)['"]`)
)

// ScanImports returns the bare package names imported anywhere under dir,
// sorted and unique. Relative paths and node: built-ins are ignored.
func ScanImports(dir string) ([]string, error) {
	byFile, err := ScanFileImports(dir)
	seen := map[string]bool{}
	for _, names := range byFile {
		for _, name := range names {
			seen[name] = true
		}
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out, err
}

// ScanFileImports returns the bare package names each source file under dir
// imports, by its slash path relative to dir. Build output, node_modules,
// hidden folders and .d.ts files are left out.
func ScanFileImports(dir string) (map[string][]string, error) {
	byFile := map[string][]string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) && path != dir {
				return nil // removed while the walk ran
			}
			return err
		}
		if d.IsDir() {
			if path != dir && (skippedDirs[d.Name()] || strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !sourceExtensions[filepath.Ext(path)] || strings.HasSuffix(path, ".d.ts") {
			return nil
		}
		src, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if names := PackageNames(string(src)); len(names) > 0 {
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			byFile[filepath.ToSlash(rel)] = names
		}
		return nil
	})
	return byFile, err
}

// PackageNames extracts bare package names from one source file's imports.
// Comments are skipped: a doc comment quoting `require('x')` imports nothing.
// So is text inside strings: in `'no "from" is given'` the from is a word,
// not a keyword. A match counts only when its keyword is code and its
// specifier is a string literal of its own.
func PackageNames(src string) []string {
	code := withoutComments(src)
	literals := stringLiterals(code)
	var out []string
	for _, m := range specifierPattern.FindAllStringSubmatchIndex(code, -1) {
		quote := m[2] - 1
		if insideLiteral(literals, m[0]) || literals[quote] != m[3] {
			continue
		}
		if name := packageName(code[m[2]:m[3]]); name != "" {
			out = append(out, name)
		}
	}
	return out
}

// stringLiterals maps the offset of each string or template literal's
// opening quote to the offset of its closing one.
func stringLiterals(code string) map[int]int {
	literals := map[int]int{}
	for i := 0; i < len(code); i++ {
		c := code[i]
		if c != '\'' && c != '"' && c != '`' {
			continue
		}
		end := i + 1
		for end < len(code) && code[end] != c && (c == '`' || code[end] != '\n') {
			if code[end] == '\\' {
				end++
			}
			end++
		}
		literals[i] = end
		i = end
	}
	return literals
}

// insideLiteral reports whether an offset falls within a literal's quotes.
func insideLiteral(literals map[int]int, at int) bool {
	for start, end := range literals {
		if at > start && at <= end {
			return true
		}
	}
	return false
}

// withoutComments blanks JavaScript's // and /* */ comments, leaving string
// and template literals (where // is part of a URL) as they are. Each comment
// becomes a space, so neither lines nor tokens join across it.
func withoutComments(src string) string {
	var b strings.Builder
	b.Grow(len(src))
	for i := 0; i < len(src); i++ {
		c := src[i]
		switch {
		case c == '\'' || c == '"' || c == '`':
			end := i + 1
			for end < len(src) && src[end] != c && (c == '`' || src[end] != '\n') {
				if src[end] == '\\' {
					end++
				}
				end++
			}
			if end >= len(src) {
				end = len(src) - 1
			}
			b.WriteString(src[i : end+1])
			i = end
		case c == '/' && i+1 < len(src) && src[i+1] == '/':
			for i < len(src) && src[i] != '\n' {
				i++
			}
			b.WriteByte(' ')
			if i < len(src) {
				b.WriteByte('\n')
			}
		case c == '/' && i+1 < len(src) && src[i+1] == '*':
			end := strings.Index(src[i+2:], "*/")
			if end < 0 {
				return b.String()
			}
			b.WriteByte(' ')
			i += end + 3
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

func packageName(specifier string) string {
	// "#x" is one of the package's own subpath imports (package.json
	// "imports"), not another package.
	if specifier == "" || strings.HasPrefix(specifier, ".") || strings.HasPrefix(specifier, "/") || strings.HasPrefix(specifier, "#") || strings.Contains(specifier, ":") {
		return ""
	}
	parts := strings.Split(specifier, "/")
	if strings.HasPrefix(specifier, "@") {
		if len(parts) < 2 {
			return ""
		}
		return parts[0] + "/" + parts[1]
	}
	return parts[0]
}
