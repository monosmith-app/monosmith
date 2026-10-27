package taskfolding

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

var tableHeader = regexp.MustCompile(`^\s*\[\[?\s*([^\]]+?)\s*\]\]?\s*(#.*)?$`)

// section is a table of smith.toml: its header line and the lines up to the
// next header.
type section struct {
	key        []string
	start, end int
}

// Rewrite applies a folded result to smith.toml text: the project task tables
// that moved into a group are removed (with the comment lines right above
// them), a table that keeps its own settings gains the group's other fields,
// and the groups are added at the end. Every other line, comments included,
// stays. ok is false, and the text is not to be written, when the file does
// not look the way folding expects (a task set in a form other than its own
// [projects.<name>.tasks.<task>] table) or when the rewritten file would not
// resolve to the same tasks.
func Rewrite(src string, projects []Project, res Result) (text string, ok bool) {
	lines := strings.SplitAfter(src, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	sections := sectionsOf(lines)
	remove := map[int]bool{}
	appendAt := map[int][]string{}
	for _, p := range projects {
		for task, original := range p.Overrides {
			s, found := find(sections, "projects", p.Name, "tasks", task)
			if !found {
				return "", false
			}
			now, kept := res.Overrides[p.Name][task]
			switch {
			case !kept:
				from := s.start
				for from > 0 && strings.HasPrefix(strings.TrimSpace(lines[from-1]), "#") {
					from--
				}
				for i := from; i < s.end; i++ {
					remove[i] = true
				}
			case !reflect.DeepEqual(now, original):
				last := s.end - 1
				for last > s.start && strings.TrimSpace(lines[last]) == "" {
					last--
				}
				appendAt[last] = workspace.TaskFieldLines(missing(original, now))
			}
		}
	}
	var b strings.Builder
	for i, line := range lines {
		if remove[i] {
			continue
		}
		b.WriteString(line)
		if extra := appendAt[i]; len(extra) > 0 {
			if !strings.HasSuffix(line, "\n") {
				b.WriteString("\n")
			}
			b.WriteString(strings.Join(extra, "\n") + "\n")
		}
	}
	out := b.String()
	if out != "" && !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	for _, g := range res.Groups {
		out += workspace.TaskGroupText(g)
	}
	return out, resolvesTheSame(src, out, projects, res)
}

// resolvesTheSame parses both texts and compares what the rewrite promised:
// every project's own task overrides as folded, and the groups as the old
// ones followed by the new.
func resolvesTheSame(before, after string, projects []Project, res Result) bool {
	old, err := workspace.Parse([]byte(before))
	if err != nil {
		return false
	}
	now, err := workspace.Parse([]byte(after))
	if err != nil {
		return false
	}
	if !reflect.DeepEqual(append(old.TaskGroups[:len(old.TaskGroups):len(old.TaskGroups)], res.Groups...), now.TaskGroups) {
		return false
	}
	names := map[string]bool{}
	for _, p := range projects {
		names[p.Name] = true
	}
	for name := range now.Projects {
		names[name] = true
	}
	for name := range names {
		want := res.Overrides[name]
		got := now.Projects[name].Tasks
		if len(want) == 0 && len(got) == 0 {
			continue
		}
		if !reflect.DeepEqual(want, got) {
			return false
		}
	}
	return true
}

func sectionsOf(lines []string) []section {
	var out []section
	for i, line := range lines {
		m := tableHeader.FindStringSubmatch(strings.TrimRight(line, "\r\n"))
		if m == nil {
			continue
		}
		if n := len(out); n > 0 {
			out[n-1].end = i
		}
		out = append(out, section{key: splitKey(m[1]), start: i, end: len(lines)})
	}
	return out
}

func find(sections []section, key ...string) (section, bool) {
	for _, s := range sections {
		if reflect.DeepEqual(s.key, key) {
			return s, true
		}
	}
	return section{}, false
}

// splitKey splits a dotted table key, taking quoted parts whole.
func splitKey(key string) []string {
	var parts []string
	var cur strings.Builder
	quote := byte(0)
	for i := 0; i < len(key); i++ {
		c := key[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			} else {
				cur.WriteByte(c)
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '.':
			parts = append(parts, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteByte(c)
		}
	}
	return append(parts, strings.TrimSpace(cur.String()))
}

// missing is the fields completed has that own does not set.
func missing(own, completed workspace.TaskOverride) workspace.TaskOverride {
	var m workspace.TaskOverride
	if own.Command == nil {
		m.Command = completed.Command
	}
	if own.Cwd == "" {
		m.Cwd = completed.Cwd
	}
	if own.Inputs == nil {
		m.Inputs = completed.Inputs
	}
	if own.Outputs == nil {
		m.Outputs = completed.Outputs
	}
	if own.DependsOn == nil {
		m.DependsOn = completed.DependsOn
	}
	if own.Env == nil {
		m.Env = completed.Env
	}
	if own.Cache == nil {
		m.Cache = completed.Cache
	}
	if own.Verify == nil {
		m.Verify = completed.Verify
	}
	if own.Serve == "" {
		m.Serve = completed.Serve
	}
	if own.ServeURL == "" {
		m.ServeURL = completed.ServeURL
	}
	if own.ServeTimeout == "" {
		m.ServeTimeout = completed.ServeTimeout
	}
	return m
}
