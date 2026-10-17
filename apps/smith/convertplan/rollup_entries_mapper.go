package convertplan

import (
	"path"
	"regexp"
	"strings"
)

// RollupEntry is one entry a library's rollup config builds: its output name
// and its source, relative to the project.
type RollupEntry struct {
	Name, Source string
}

var (
	rollupInputKey      = regexp.MustCompile(`\binput\s*:\s*`)
	rollupAdditionalKey = regexp.MustCompile(`\badditionalEntryPoints\s*:\s*`)
	rollupEntryName     = regexp.MustCompile(`^['"]?([A-Za-z_$][\w$-]*)['"]?$`)
	// A source as a rollup config writes it: a string, a template string on
	// __dirname, or path.join/resolve(__dirname, ...).
	rollupQuoted   = regexp.MustCompile(`^['"]([^'"]+)['"]$`)
	rollupTemplate = regexp.MustCompile("^`\\$\\{__dirname\\}/([^`$]+)`$")
	rollupJoin     = regexp.MustCompile(`^(?:path\.)?(?:join|resolve)\(\s*__dirname\s*,\s*['"]([^'"]+)['"]\s*\)$`)
)

// RollupEntries reads the entries a rollup config builds beyond `main`:
// Rollup's own `input` (an object of named entries, or a list named after
// their files) or withNx's `additionalEntryPoints` (named after their
// files). It returns nil when the config builds main alone, and ok false
// when it names entries convert cannot read, so the conversion asks instead
// of building main only.
func RollupEntries(text string) (entries []RollupEntry, ok bool) {
	text = stripLineComments(text)
	key := rollupInputKey.FindStringIndex(text)
	if key == nil {
		key = rollupAdditionalKey.FindStringIndex(text)
	}
	if key == nil {
		return nil, true
	}
	body, open, ok := bracketed(text[key[1]:])
	if !ok {
		return nil, false
	}
	seen := map[string]bool{}
	for _, item := range splitTopLevel(body) {
		var name, expr string
		if open == '{' {
			k, v, found := strings.Cut(item, ":")
			m := rollupEntryName.FindStringSubmatch(strings.TrimSpace(k))
			if !found || m == nil {
				return nil, false
			}
			name, expr = m[1], v
		} else {
			expr = item
		}
		src, ok := rollupSource(expr)
		if !ok {
			return nil, false
		}
		if name == "" {
			// Rollup names a listed entry after its file.
			name = strings.TrimSuffix(path.Base(src), path.Ext(src))
		}
		if seen[name] {
			return nil, false
		}
		seen[name] = true
		entries = append(entries, RollupEntry{Name: name, Source: src})
	}
	return entries, len(entries) > 0
}

// bracketed returns what the object or list starting text holds, skipping
// strings and template literals, whose braces do not count.
func bracketed(text string) (body string, open byte, ok bool) {
	if text == "" || (text[0] != '{' && text[0] != '[') {
		return "", 0, false
	}
	open = text[0]
	depth := 0
	var quote byte
	for i := 0; i < len(text); i++ {
		c := text[i]
		switch {
		case quote != 0:
			switch c {
			case '\\':
				i++
			case quote:
				quote = 0
			}
		case c == '\'' || c == '"' || c == '`':
			quote = c
		case c == '{' || c == '[' || c == '(':
			depth++
		case c == '}' || c == ']' || c == ')':
			depth--
			if depth == 0 {
				return text[1:i], open, true
			}
		}
	}
	return "", 0, false
}

// splitTopLevel splits a list on its commas outside strings and brackets.
func splitTopLevel(body string) []string {
	var items []string
	depth, start := 0, 0
	var quote byte
	for i := 0; i < len(body); i++ {
		c := body[i]
		switch {
		case quote != 0:
			switch c {
			case '\\':
				i++
			case quote:
				quote = 0
			}
		case c == '\'' || c == '"' || c == '`':
			quote = c
		case c == '{' || c == '[' || c == '(':
			depth++
		case c == '}' || c == ']' || c == ')':
			depth--
		case c == ',' && depth == 0:
			items = append(items, body[start:i])
			start = i + 1
		}
	}
	items = append(items, body[start:])
	var out []string
	for _, it := range items {
		if it = strings.TrimSpace(it); it != "" {
			out = append(out, it)
		}
	}
	return out
}

func rollupSource(expr string) (string, bool) {
	expr = strings.TrimSpace(expr)
	for _, re := range []*regexp.Regexp{rollupQuoted, rollupTemplate, rollupJoin} {
		if m := re.FindStringSubmatch(expr); m != nil {
			src := path.Clean(strings.TrimPrefix(m[1], "./"))
			if strings.HasPrefix(src, "../") || !plainFlagValue(src) {
				return "", false
			}
			return src, true
		}
	}
	return "", false
}

var (
	// lineComment is a // comment that does not start inside a URL (://).
	lineComment  = regexp.MustCompile(`(?m)(^|[^:])//[^\n]*$`)
	blockComment = regexp.MustCompile(`(?s)/\*.*?\*/`)
)

// stripLineComments drops comments, whose apostrophes would read as quotes.
func stripLineComments(s string) string {
	return lineComment.ReplaceAllString(blockComment.ReplaceAllString(s, ""), "$1")
}
