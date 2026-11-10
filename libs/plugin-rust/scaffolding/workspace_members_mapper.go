package scaffolding

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	workspaceHeader = regexp.MustCompile(`(?m)^\[workspace\][ \t]*(#.*)?\r?\n?`)
	quotedItem      = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)
)

// NewWorkspace is the root Cargo.toml of a workspace that has only members.
func NewWorkspace(member string) string {
	return "[workspace]\nresolver = \"2\"\nmembers = [\"" + member + "\"]\n"
}

// members finds the members array of the [workspace] table: where it sits,
// what it lists, and whether it spans lines.
func members(src string) (start, end int, items []string, multiline bool, err error) {
	return keyList(src, "members")
}

// keyList is members for any array key of the [workspace] table ("members",
// "exclude").
func keyList(src, key string) (start, end int, items []string, multiline bool, err error) {
	keyStart := regexp.MustCompile(`(?m)^[ \t]*` + regexp.QuoteMeta(key) + `[ \t]*=[ \t]*\[`)
	h := workspaceHeader.FindStringIndex(src)
	if h == nil {
		return 0, 0, nil, false, fmt.Errorf("the root Cargo.toml has no [workspace] table")
	}
	body := src[h[1]:]
	if next := regexp.MustCompile(`(?m)^\[`).FindStringIndex(body); next != nil {
		body = body[:next[0]]
	}
	m := keyStart.FindStringIndex(body)
	if m == nil {
		return h[1], h[1], nil, false, nil
	}
	start = h[1] + m[1] - 1
	closing := strings.Index(src[start:], "]")
	if closing < 0 {
		return 0, 0, nil, false, fmt.Errorf("the root Cargo.toml's members are not closed")
	}
	end = start + closing + 1
	for _, it := range quotedItem.FindAllStringSubmatch(src[start:end], -1) {
		items = append(items, it[1])
	}
	return start, end, items, strings.Contains(src[start:end], "\n"), nil
}

func render(items []string, multiline bool) string {
	quoted := make([]string, len(items))
	for i, it := range items {
		quoted[i] = `"` + it + `"`
	}
	if multiline && len(items) > 0 {
		return "[\n    " + strings.Join(quoted, ",\n    ") + ",\n]"
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

// AddMember lists a crate directory in the workspace's members; listing it
// again changes nothing.
func AddMember(src, member string) (string, error) { return addToList(src, "members", member) }

// AddExclude lists a directory in the workspace's exclude array, so cargo
// leaves a crate there out of the workspace; listing it again changes nothing.
func AddExclude(src, dir string) (string, error) { return addToList(src, "exclude", dir) }

func addToList(src, key, member string) (string, error) {
	start, end, items, multiline, err := keyList(src, key)
	if err != nil {
		return "", err
	}
	for _, it := range items {
		if it == member {
			return src, nil
		}
	}
	items = append(items, member)
	if start == end { // no members key yet
		return src[:start] + key + " = " + render(items, false) + "\n" + src[start:], nil
	}
	return src[:start] + render(items, multiline) + src[end:], nil
}

// DropMember takes a crate directory out of the members; changed is false
// when it was not listed.
func DropMember(src, member string) (string, bool) {
	start, end, items, multiline, err := members(src)
	if err != nil || start == end {
		return src, false
	}
	var kept []string
	for _, it := range items {
		if it != member {
			kept = append(kept, it)
		}
	}
	if len(kept) == len(items) {
		return src, false
	}
	return src[:start] + render(kept, multiline) + src[end:], true
}
