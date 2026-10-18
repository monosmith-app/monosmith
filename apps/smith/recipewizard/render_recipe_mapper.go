package recipewizard

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var bareKey = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// Render writes a plan as the recipe a person would write: the choices, the
// plugins that are fixed under [use], and for each answer a [when] block with
// the plugins it picks.
func Render(p Plan) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Recipe %s, created with `smith recipe new`.\n", p.Name)
	b.WriteString("# Edit it by hand if you like; `smith recipe check <this file> --all` checks it\n")
	b.WriteString("# for every answer.\n")
	fmt.Fprintf(&b, "name = %s\nversion = %s\n", quote(p.Name), quote(p.Version))
	if p.Extends != "" {
		fmt.Fprintf(&b, "extends = %s\n", quote(p.Extends))
	}
	for _, name := range keys(p.Choices) {
		c := p.Choices[name]
		fmt.Fprintf(&b, "\n[choices.%s]\n", key(name))
		if c.Ask != "" {
			fmt.Fprintf(&b, "ask = %s\n", quote(c.Ask))
		}
		fmt.Fprintf(&b, "options = %s\n", list(c.Options))
		if c.Default != "" {
			fmt.Fprintf(&b, "default = %s\n", quote(c.Default))
		}
	}
	levels := OrderedLevels(p.Levels)
	var fixed []string
	for _, level := range levels {
		if len(p.Levels[level].Fixed) > 0 {
			fixed = append(fixed, level)
		}
	}
	if len(fixed) > 0 {
		b.WriteString("\n[use]\n")
		for _, level := range fixed {
			fmt.Fprintf(&b, "%s = %s\n", key(level), plugins(p.Levels[level].Fixed))
		}
	}
	for _, choice := range choicesUsed(p) {
		for _, answer := range answersOf(p, choice) {
			var lines []string
			for _, level := range levels {
				l := p.Levels[level]
				if l.Choice != choice {
					continue
				}
				if names := l.PluginsOf(answer); len(names) > 0 {
					lines = append(lines, fmt.Sprintf("%s = %s\n", key(level), plugins(names)))
				}
			}
			if len(lines) == 0 {
				continue
			}
			fmt.Fprintf(&b, "\n[when.%s.%s.use]\n%s", key(choice), key(answer), strings.Join(lines, ""))
		}
	}
	return b.String()
}

// choicesUsed lists the choices some level is picked by: those the plan
// declares, then those of the parent it only refers to, each sorted.
func choicesUsed(p Plan) []string {
	seen := map[string]bool{}
	for _, l := range p.Levels {
		if l.Choice != "" {
			seen[l.Choice] = true
		}
	}
	return keys(seen)
}

// answersOf lists a choice's answers in the order its options give, else the
// sorted answers the levels name (a choice the parent declares).
func answersOf(p Plan, choice string) []string {
	if c, ok := p.Choices[choice]; ok {
		return c.Options
	}
	seen := map[string]bool{}
	for _, l := range p.Levels {
		if l.Choice == choice {
			for answer := range l.Plugins {
				seen[answer] = true
			}
		}
	}
	out := keys(seen)
	sort.Strings(out)
	return out
}

func plugins(names []string) string {
	if len(names) == 1 {
		return quote(names[0])
	}
	return list(names)
}

func list(values []string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = quote(v)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// key writes a TOML key, quoted unless it is a bare one.
func key(s string) string {
	if bareKey.MatchString(s) {
		return s
	}
	return quote(s)
}

// quote writes a TOML basic string.
func quote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\t':
			b.WriteString(`\t`)
		case r < 0x20 || r == 0x7f:
			fmt.Fprintf(&b, `\u%04X`, r)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
