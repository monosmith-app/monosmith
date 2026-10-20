package toolchainpin

import (
	"sort"
	"strings"
)

// toolNames maps a toolchain label to the name mise and asdf know it by,
// where they differ.
var toolNames = map[string]string{"rustc": "rust"}

// ToolName is the name .tool-versions uses for a toolchain label.
func ToolName(label string) string {
	if n, ok := toolNames[label]; ok {
		return n
	}
	return label
}

// MergeToolVersions returns the .tool-versions text with a line for every
// pin: a tool already listed gets the pinned version, other lines (and
// comments) stay as they are, and new tools are appended in name order.
func MergeToolVersions(existing string, pins map[string]string) string {
	want := map[string]string{}
	for label, v := range pins {
		want[ToolName(label)] = strings.TrimPrefix(v, "v")
	}
	var out []string
	if existing != "" {
		out = strings.Split(strings.TrimRight(strings.ReplaceAll(existing, "\r\n", "\n"), "\n"), "\n")
	}
	seen := map[string]bool{}
	for i, line := range out {
		fields := strings.Fields(line)
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		if v, ok := want[fields[0]]; ok {
			out[i] = fields[0] + " " + v
			seen[fields[0]] = true
		}
	}
	var added []string
	for tool := range want {
		if !seen[tool] {
			added = append(added, tool)
		}
	}
	sort.Strings(added)
	for _, tool := range added {
		out = append(out, tool+" "+want[tool])
	}
	return strings.Join(out, "\n") + "\n"
}
