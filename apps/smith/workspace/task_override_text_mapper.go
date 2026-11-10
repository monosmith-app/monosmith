package workspace

import (
	"fmt"
	"sort"
	"strings"
)

// TaskFieldLines writes the fields a task override sets as smith.toml lines
// (key = value, no newline), in a stable order.
func TaskFieldLines(t TaskOverride) []string {
	var out []string
	add := func(format string, args ...any) { out = append(out, fmt.Sprintf(format, args...)) }
	if t.Command != nil {
		add("command = %s", tomlList(t.Command))
	}
	if t.Cwd != "" {
		add("cwd = %s", tomlString(t.Cwd))
	}
	if t.DependsOn != nil {
		add("depends_on = %s", tomlList(t.DependsOn))
	}
	if t.Inputs != nil {
		add("inputs = %s", tomlList(t.Inputs))
	}
	if t.Outputs != nil {
		add("outputs = %s", tomlList(t.Outputs))
	}
	if t.Env != nil {
		add("env = %s", tomlInlineTable(t.Env))
	}
	if t.Cache != nil {
		add("cache = %t", *t.Cache)
	}
	if t.Verify != nil {
		add("verify = %t", *t.Verify)
	}
	if t.Serve != "" {
		add("serve = %s", tomlString(t.Serve))
	}
	if t.ServeURL != "" {
		add("serve_url = %s", tomlString(t.ServeURL))
	}
	if t.ServeTimeout != "" {
		add("serve_timeout = %s", tomlString(t.ServeTimeout))
	}
	return out
}

// GroupText writes one [[groups]] entry with its task tables, ending
// in a newline and starting with a blank line.
func GroupText(g Group) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n[[groups]]\nselect = %s\n", tomlList(g.Select))
	names := make([]string, 0, len(g.Tasks))
	for name := range g.Tasks {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(&b, "\n[groups.tasks.%s]\n", tomlKey(name))
		for _, line := range TaskFieldLines(g.Tasks[name]) {
			b.WriteString(line + "\n")
		}
	}
	return b.String()
}

func tomlString(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\t", `\t`, "\r", `\r`)
	return `"` + r.Replace(s) + `"`
}

func tomlKey(s string) string {
	for _, c := range s {
		bare := c == '-' || c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
		if !bare {
			return tomlString(s)
		}
	}
	return s
}

func tomlList(items []string) string {
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = tomlString(s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func tomlInlineTable(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = tomlKey(k) + " = " + tomlString(m[k])
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}
