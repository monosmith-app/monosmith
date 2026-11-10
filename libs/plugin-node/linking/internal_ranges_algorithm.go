package linking

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
)

// AnySpec is the range smith writes when a project depends on a package of its
// own workspace: npm resolves it from the workspace whatever its version.
const AnySpec = "*"

// rangeSections are the sections whose range is a request to install: a
// peerDependencies range is a contract for consumers and is left alone.
var rangeSections = []string{"dependencies", "devDependencies", "optionalDependencies"}

// PinnedRange is a dependency on a package of the workspace with a range other
// than AnySpec.
type PinnedRange struct {
	// Section is the manifest section that declares the entry.
	Section string
	// Name is the package the entry depends on.
	Name string
	// Spec is the range as written.
	Spec string
	// start and end bound the quoted range in the manifest text.
	start, end int
}

// String names the entry: the package, its range and the section.
func (p PinnedRange) String() string {
	return p.Name + " " + p.Spec + " (" + p.Section + ")"
}

// isPinned reports whether spec is a registry range. A workspace:, file:,
// link:, git or URL spec says where to get the package, so it is not pinned.
func isPinned(spec string) bool {
	switch {
	case spec == AnySpec:
		return false
	case strings.HasPrefix(spec, "workspace:"), strings.HasPrefix(spec, "file:"), strings.HasPrefix(spec, "link:"),
		strings.HasPrefix(spec, "git"), strings.HasPrefix(spec, "github:"), strings.HasPrefix(spec, "npm:"),
		strings.Contains(spec, "://"), strings.Contains(spec, "/"):
		return false
	}
	return true
}

// PinnedInternalRanges lists the entries of dependencies, devDependencies and
// optionalDependencies of a package.json text that name one of internal (the
// package names of the workspace) with a range other than AnySpec, in file
// order. It returns an error when the text is not a JSON object.
func PinnedInternalRanges(manifest []byte, internal map[string]bool) ([]PinnedRange, error) {
	dec := json.NewDecoder(bytes.NewReader(manifest))
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	var out []PinnedRange
	for dec.More() {
		key, err := dec.Token()
		if err != nil {
			return nil, err
		}
		section, _ := key.(string)
		if !isRangeSection(section) {
			var skip json.RawMessage
			if err := dec.Decode(&skip); err != nil {
				return nil, err
			}
			continue
		}
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		if d, ok := tok.(json.Delim); !ok || d != '{' {
			continue
		}
		for dec.More() {
			nameTok, err := dec.Token()
			if err != nil {
				return nil, err
			}
			var raw json.RawMessage
			if err := dec.Decode(&raw); err != nil {
				return nil, err
			}
			name, _ := nameTok.(string)
			var spec string
			if !internal[name] || json.Unmarshal(raw, &spec) != nil || !isPinned(spec) {
				continue
			}
			end := int(dec.InputOffset())
			out = append(out, PinnedRange{Section: section, Name: name, Spec: spec, start: end - len(raw), end: end})
		}
		if _, err := dec.Token(); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func isRangeSection(s string) bool {
	for _, r := range rangeSections {
		if s == r {
			return true
		}
	}
	return false
}

// UseAnySpec returns the manifest text with the range of each pinned entry
// replaced by AnySpec. Only those quoted values change: the indentation, the
// line endings, the key order and every other byte stay as they were. The
// pinned entries must come from PinnedInternalRanges on the same text.
func UseAnySpec(manifest []byte, pinned []PinnedRange) []byte {
	ordered := append([]PinnedRange(nil), pinned...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].start > ordered[j].start })
	out := append([]byte(nil), manifest...)
	for _, p := range ordered {
		tail := append([]byte(`"`+AnySpec+`"`), out[p.end:]...)
		out = append(out[:p.start], tail...)
	}
	return out
}
