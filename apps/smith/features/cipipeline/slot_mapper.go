package cipipeline

import (
	"regexp"
	"strings"
)

var slotPattern = regexp.MustCompile(`(?ms)^([ \t]*)# smith:slot ([\w-]+)[ \t]*\r?\n(.*?)^[ \t]*# smith:slot-end ([\w-]+)[ \t]*\r?$`)

// KeepSlots copies the content between each smith:slot marker pair of the old
// file into the same slot of the newly rendered one, re-indented when the
// slot moved (an Azure pipeline gaining a native job nests its steps under
// jobs). A slot the new file no longer has is reported, so a team step is
// never dropped silently.
func KeepSlots(old, rendered string) (string, []string) {
	bodies := map[string]string{}
	indents := map[string]string{}
	for _, m := range slotPattern.FindAllStringSubmatch(old, -1) {
		if m[2] == m[4] && strings.TrimSpace(m[3]) != "" {
			bodies[m[2]], indents[m[2]] = m[3], m[1]
		}
	}
	used := map[string]bool{}
	out := slotPattern.ReplaceAllStringFunc(rendered, func(block string) string {
		m := slotPattern.FindStringSubmatch(block)
		body, ok := bodies[m[2]]
		if !ok {
			return block
		}
		used[m[2]] = true
		if from := indents[m[2]]; from != m[1] {
			body = reindent(body, from, m[1])
		}
		return m[1] + "# smith:slot " + m[2] + "\n" + body + m[1] + "# smith:slot-end " + m[2]
	})
	var lost []string
	for name := range bodies {
		if !used[name] {
			lost = append(lost, name)
		}
	}
	return out, lost
}

// reindent moves every line indented by at least from to indent to.
func reindent(body, from, to string) string {
	lines := strings.SplitAfter(body, "\n")
	for i, l := range lines {
		if strings.HasPrefix(l, from) {
			lines[i] = to + strings.TrimPrefix(l, from)
		}
	}
	return strings.Join(lines, "")
}
