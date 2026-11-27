package palette

import "strings"

// commandLine is the smith command line a hint starts with: the part before
// ": " when the hint is "smith <command>: what it does", else the whole hint
// (a hint that is no command line is shown as it is).
func commandLine(hint string) string {
	if !strings.HasPrefix(hint, "smith ") { // words:literal
		return hint
	}
	if cmd, _, ok := strings.Cut(hint, ": "); ok {
		return cmd
	}
	return hint
}

// matchSpans are the byte ranges of title that the typed query matches: the
// whole query where the title holds it, else each typed word the title holds,
// in order and without overlap. The search ignores case.
func matchSpans(title, query string) [][2]int {
	q := strings.ToLower(strings.TrimSpace(query))
	lower := strings.ToLower(title)
	if q == "" || len(lower) != len(title) {
		// A letter whose lower case is longer would move every span: no highlight.
		return nil
	}
	if at := strings.Index(lower, q); at >= 0 {
		return [][2]int{{at, at + len(q)}}
	}
	var spans [][2]int
	from := 0
	for _, w := range strings.Fields(q) {
		at := strings.Index(lower[from:], w)
		if at < 0 {
			continue
		}
		spans = append(spans, [2]int{from + at, from + at + len(w)})
		from += at + len(w)
	}
	return spans
}

// highlight draws title with each span in lit and the rest in plain.
func highlight(title string, spans [][2]int, plain, lit func(string) string) string {
	var b strings.Builder
	at := 0
	for _, s := range spans {
		if s[0] > at {
			b.WriteString(plain(title[at:s[0]]))
		}
		b.WriteString(lit(title[s[0]:s[1]]))
		at = s[1]
	}
	if at < len(title) {
		b.WriteString(plain(title[at:]))
	}
	return b.String()
}
