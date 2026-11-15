package release

import (
	"fmt"
	"sort"
	"strings"
)

// Change is one releasing commit, as release notes show it.
type Change struct {
	Type        string   `json:"type"`
	Scopes      []string `json:"scopes,omitempty"`
	Description string   `json:"description"`
	Breaking    bool     `json:"breaking,omitempty"`
	// BreakingNote is what the BREAKING CHANGE footer says, when it says it.
	BreakingNote string `json:"breakingNote,omitempty"`
	Hash         string `json:"hash"`
	Author       string `json:"author,omitempty"`
}

// sections are the release notes' headings for each conventional type, in
// order; other types do not release, so they are not listed.
var sections = []struct{ kind, title string }{
	{"feat", "Features"},
	{"fix", "Fixes"},
	{"perf", "Performance"},
	{"revert", "Reverts"},
}

// RenderNotes writes an item's release notes in Markdown: breaking changes
// first, then features, fixes, performance and reverts, a dependency release
// when that is the reason, and thanks to the people whose commits it holds.
func RenderNotes(item Item) string {
	var b strings.Builder
	var breaking []Change
	for _, c := range item.Changes {
		if c.Breaking {
			breaking = append(breaking, c)
		}
	}
	if len(breaking) > 0 {
		b.WriteString("## ▲ Breaking changes\n\n")
		for _, c := range breaking {
			line := entry(c)
			if c.BreakingNote != "" {
				line += "\n  " + strings.ReplaceAll(c.BreakingNote, "\n", "\n  ")
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}
	for _, s := range sections {
		var lines []string
		for _, c := range item.Changes {
			if c.Type == s.kind {
				lines = append(lines, entry(c))
			}
		}
		if len(lines) > 0 {
			fmt.Fprintf(&b, "## %s\n\n%s\n\n", s.title, strings.Join(lines, "\n"))
		}
	}
	if len(item.Changes) == 0 {
		fmt.Fprintf(&b, "## Dependencies\n\n- %s\n\n", item.Reason)
	}
	if people := thanks(item.Changes); len(people) > 0 {
		fmt.Fprintf(&b, "## Thanks\n\n%s\n", sentence(people))
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func entry(c Change) string {
	line := "- "
	if len(c.Scopes) > 0 {
		line += "**" + strings.Join(c.Scopes, ", ") + ":** "
	}
	line += c.Description
	if len(c.Hash) >= 7 {
		line += " (" + c.Hash[:7] + ")"
	}
	return line
}

// thanks are the distinct authors, bots aside, in alphabetical order.
func thanks(changes []Change) []string {
	seen := map[string]bool{}
	var out []string
	for _, c := range changes {
		a := strings.TrimSpace(c.Author)
		lower := strings.ToLower(a)
		if a == "" || seen[a] || strings.HasSuffix(lower, "[bot]") || strings.Contains(lower, "dependabot") || strings.Contains(lower, "github-actions") {
			continue
		}
		seen[a] = true
		out = append(out, a)
	}
	sort.Strings(out)
	return out
}

func sentence(people []string) string {
	switch len(people) {
	case 1:
		return "Thanks to " + people[0] + "."
	case 2:
		return "Thanks to " + people[0] + " and " + people[1] + "."
	}
	return "Thanks to " + strings.Join(people[:len(people)-1], ", ") + " and " + people[len(people)-1] + "."
}
