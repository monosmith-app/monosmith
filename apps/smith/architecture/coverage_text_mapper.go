package architecture

import (
	"strconv"
	"strings"
)

// shown is how many projects a line names before it counts the rest.
const shown = 4

// AppliesText is the languages the architecture applies to, in a line: "go,
// node, python", or "every language" when its plugin names none. It is empty
// for the choice of none, which applies to nothing.
func (a Architecture) AppliesText() string {
	if a.Table == "" {
		return ""
	}
	if languages := a.AppliesTo(); len(languages) > 0 {
		return strings.Join(languages, ", ")
	}
	return "every language"
}

// Here says what the architecture does to this workspace's projects: which it
// checks and which it skips because of their language. It is empty when the
// workspace has no projects, or the choice is none.
func (c Coverage) Here() string {
	var parts []string
	if len(c.Checked) > 0 {
		parts = append(parts, "checks "+list(c.Checked))
	}
	if len(c.Skipped) > 0 {
		parts = append(parts, "skips "+list(c.Skipped)+" (its language is not covered)")
	}
	return strings.Join(parts, "; ")
}

// Warning is what to say before the architecture is chosen when it would check
// none of the workspace's projects; empty when it would check at least one.
func (a Architecture) Warning(c Coverage) string {
	if !c.CoversNone() {
		return ""
	}
	return a.Name + " applies to " + a.AppliesText() + ", and none of your projects is written in those languages (" + list(c.Skipped) + "), so it would check nothing."
}

// list names members, a few at most, as "api (go), web (node)".
func list(ms []Member) string {
	var out []string
	for i, m := range ms {
		if i == shown {
			out = append(out, "and "+strconv.Itoa(len(ms)-shown)+" more")
			break
		}
		out = append(out, label(m))
	}
	return strings.Join(out, ", ")
}

func label(m Member) string {
	if m.Language == "" {
		return m.Name
	}
	return m.Name + " (" + m.Language + ")"
}
