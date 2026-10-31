package apisurface

import (
	"regexp"
	"sort"
	"strings"
)

// Exports is what one module file exports by itself.
type Exports struct {
	// Names are the exported names the file declares or re-exports by name.
	Names []string
	// Stars are the specifiers of "export * from": the names of those modules
	// are exported too, and the caller follows them.
	Stars []string
}

var (
	namedExportRe  = regexp.MustCompile(`^\s*(?:type\s+)?\{([^}]*)\}(?:\s*from\s*(['"][^'"]*['"]))?`)
	starExportRe   = regexp.MustCompile(`^\s*\*\s*(?:as\s+([\w$]+)\s+)?from\s*(['"][^'"]*['"])`)
	assignExportRe = regexp.MustCompile(`^\s*=\s*[\w$.]+`)
	defaultRe      = regexp.MustCompile(`^\s+default\b`)
	importEqualsRe = regexp.MustCompile(`^\s+import\s+([\w$]+)\s*=`)
	declarationRe  = regexp.MustCompile(`^\s+(?:declare\s+)?(?:abstract\s+)?(?:async\s+)?(?:function\s*\*?|class|const\s+enum|enum|interface|type|namespace|module|let|const|var)\s*([\w$]+)`)
	listItemRe     = regexp.MustCompile(`^(?:type\s+)?([\w$]+)(?:\s+as\s+([\w$]+))?$`)

	commonJSMemberRe = regexp.MustCompile(`\b(?:module\.)?exports\.([\w$]+)\s*=[^=]`)
	commonJSObjectRe = regexp.MustCompile(`\bmodule\.exports\s*=\s*\{([^}]*)\}`)
	commonJSWholeRe  = regexp.MustCompile(`\bmodule\.exports\s*=[^=]`)
	commonJSDefineRe = regexp.MustCompile(`Object\.defineProperty\(\s*(?:module\.)?exports\s*,\s*['"]([\w$]+)['"]`)
)

// ParseExports reads the exports a JavaScript or TypeScript file declares:
// ES export statements at the top level (not inside a namespace, class or
// function body) and, outside declaration files, the CommonJS forms
// exports.name =, module.exports = { ... } and Object.defineProperty(exports,
// "name"). A file that is not valid code gives whatever the scanner finds.
func ParseExports(file, src string) Exports {
	plain, masked := maskSource(src)
	var out Exports
	seen := map[string]bool{}
	add := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			out.Names = append(out.Names, name)
		}
	}
	scanExportStatements(string(plain), string(masked), add, &out)
	if !strings.HasSuffix(file, ".d.ts") {
		scanCommonJS(string(plain), add)
	}
	sort.Strings(out.Names)
	return out
}

// scanExportStatements walks the top level of the file and reads each export
// statement. masked has the same length as plain, with comments and the
// content of strings blanked, so a brace or a keyword inside either is never
// seen.
func scanExportStatements(plain, masked string, add func(string), out *Exports) {
	depth := 0
	for i := 0; i < len(masked); {
		switch c := masked[i]; {
		case c == '{':
			depth++
		case c == '}':
			depth = max(0, depth-1)
		case depth == 0 && c == 'e' && isExportKeyword(masked, i):
			i += scanOneExport(plain, masked, i+len("export"), add, out)
			continue
		}
		i++
	}
}

// isExportKeyword is whether the word "export" starts at i.
func isExportKeyword(s string, i int) bool {
	if !strings.HasPrefix(s[i:], "export") {
		return false
	}
	if i > 0 && isIdentChar(s[i-1]) || i > 0 && s[i-1] == '.' {
		return false
	}
	end := i + len("export")
	return end == len(s) || !isIdentChar(s[end])
}

func isIdentChar(c byte) bool {
	return c == '_' || c == '$' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// scanOneExport reads the statement after the word export, which ends at
// from. It adds what the statement exports and returns how many bytes it used
// (at least the keyword, so the caller always moves on).
func scanOneExport(plain, masked string, from int, add func(string), out *Exports) int {
	rest := masked[from:]
	used := len("export")
	if m := starExportRe.FindStringSubmatchIndex(rest); m != nil {
		if m[2] >= 0 {
			add(rest[m[2]:m[3]])
		} else {
			out.Stars = append(out.Stars, quoted(plain, from+m[4], from+m[5]))
		}
		return used + m[1]
	}
	if m := namedExportRe.FindStringSubmatchIndex(rest); m != nil {
		for _, item := range strings.Split(rest[m[2]:m[3]], ",") {
			if sub := listItemRe.FindStringSubmatch(strings.TrimSpace(item)); sub != nil {
				add(firstNonEmpty(sub[2], sub[1]))
			}
		}
		return used + m[1]
	}
	if m := assignExportRe.FindStringIndex(rest); m != nil {
		add("export=")
		return used + m[1]
	}
	if m := defaultRe.FindStringIndex(rest); m != nil {
		add("default")
		return used + m[1]
	}
	if m := importEqualsRe.FindStringSubmatch(rest); m != nil {
		add(m[1])
		return used + len(m[0])
	}
	if m := declarationRe.FindStringSubmatch(rest); m != nil {
		add(m[1])
		return used + len(m[0])
	}
	return used
}

// quoted returns the text between the quotes that sit at from and to-1 of
// the unmasked source.
func quoted(plain string, from, to int) string {
	return plain[from+1 : to-1]
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// scanCommonJS reads the CommonJS export forms, at any depth.
func scanCommonJS(plain string, add func(string)) {
	for _, m := range commonJSMemberRe.FindAllStringSubmatch(plain, -1) {
		if m[1] != "__esModule" {
			add(m[1])
		}
	}
	for _, m := range commonJSDefineRe.FindAllStringSubmatch(plain, -1) {
		if m[1] != "__esModule" {
			add(m[1])
		}
	}
	if m := commonJSObjectRe.FindStringSubmatch(plain); m != nil {
		for _, item := range strings.Split(m[1], ",") {
			key, _, _ := strings.Cut(strings.TrimSpace(item), ":")
			if key = strings.TrimSpace(key); isIdentifier(key) {
				add(key)
			}
		}
	} else if commonJSWholeRe.MatchString(plain) {
		add("module.exports")
	}
}

func isIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !isIdentChar(s[i]) {
			return false
		}
	}
	return true
}

// maskSource returns two copies of src with the same length: plain has the
// comments blanked, masked also has the content of strings and template
// literals blanked (the quotes stay). An unterminated quote ends at the line
// break, so a regular expression with a quote does not hide the rest of the
// file.
func maskSource(src string) (plain, masked []byte) {
	plain, masked = []byte(src), []byte(src)
	blank := func(i int, both bool) {
		if src[i] == '\n' || src[i] == '\r' {
			return
		}
		masked[i] = ' '
		if both {
			plain[i] = ' '
		}
	}
	n := len(src)
	for i := 0; i < n; {
		c := src[i]
		switch {
		case c == '/' && i+1 < n && src[i+1] == '/':
			for ; i < n && src[i] != '\n'; i++ {
				blank(i, true)
			}
		case c == '/' && i+1 < n && src[i+1] == '*':
			stop := n
			if end := strings.Index(src[i+2:], "*/"); end >= 0 {
				stop = i + 2 + end + 2
			}
			for ; i < stop; i++ {
				blank(i, true)
			}
		case c == '\'' || c == '"' || c == '`':
			i++
			for i < n && src[i] != c {
				if c != '`' && src[i] == '\n' {
					break
				}
				if src[i] == '\\' && i+1 < n {
					blank(i, false)
					i++
				}
				blank(i, false)
				i++
			}
			i++
		default:
			i++
		}
	}
	return plain, masked
}
