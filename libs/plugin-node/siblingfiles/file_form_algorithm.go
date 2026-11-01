package siblingfiles

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// The sentinels stand for the values that belong to one project. They are
// control characters, so they never occur in a config file.
const (
	sentinelDisplay = "\x01"
	sentinelName    = "\x02"
	sentinelRefs    = "\x03references"
)

// refsLacking is the refsAt of a file with no references although the project
// has dependencies. It equals no position.
const refsLacking = -2

// referencesKey is the tsconfig key that lists the internal dependencies.
const referencesKey = "references"

// References answers the references of a project's tsconfig: one entry for each
// internal dependency, in the order of content (the file as it is, or nil when
// there is none) with new entries last. An entry is a JSON object with a path.
type References func(p protocol.Project, content []byte) []any

// field is a top-level key of a JSON file and its compact value, with the
// project's own values replaced by sentinels.
type field struct{ key, value string }

// member is one project's file at the path under comparison.
type member struct {
	project protocol.Project
	rel     string
	raw     []byte
	exists  bool
	owned   bool
	// text is the abstracted text of a file that is not a JSON object.
	text string
	// isJSON says the file is a JSON object, described by fields.
	isJSON bool
	fields []field
	// refsAt is the number of fields before the references key when the file
	// lists exactly the project's internal dependencies, or -1.
	refsAt int
	// refsWild is true when the file lists none and the project has no
	// dependency: it fits a shape with or without the key.
	refsWild bool
	// deps is how many internal dependencies the project has.
	deps int
}

// entry is a top-level key of a JSON object with its raw value.
type entry struct {
	key string
	raw json.RawMessage
}

// topLevel reads the top-level entries of a JSON object, in file order. ok is
// false when the text is not a JSON object (comments, for one).
func topLevel(raw []byte) ([]entry, bool) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return nil, false
	}
	var out []entry
	for dec.More() {
		t, err := dec.Token()
		key, isKey := t.(string)
		if err != nil || !isKey {
			return nil, false
		}
		var v json.RawMessage
		if dec.Decode(&v) != nil {
			return nil, false
		}
		out = append(out, entry{key, v})
	}
	if _, err := dec.Token(); err != nil || dec.More() {
		return nil, false
	}
	return out, true
}

// tokens are the values of one project that a file may repeat.
type tokens struct{ display, name string }

func tokensOf(p protocol.Project) tokens {
	t := tokens{name: p.Name, display: p.Name}
	if len(p.Exports) > 0 {
		t.display = p.Exports[0]
	}
	return t
}

// abstract replaces the project's display name and name by sentinels.
func (t tokens) abstract(s string) string {
	if t.display != t.name {
		s = replaceWhole(s, t.display, sentinelDisplay)
	}
	return replaceWhole(s, t.name, sentinelName)
}

// concrete is the inverse of abstract.
func (t tokens) concrete(s string) string {
	return strings.NewReplacer(sentinelDisplay, t.display, sentinelName, t.name).Replace(s)
}

// replaceWhole replaces the occurrences of word that are not part of a longer
// name: the characters around it are not letters, digits, "_", "-", "." or "@".
func replaceWhole(s, word, with string) string {
	if word == "" {
		return s
	}
	var b strings.Builder
	for {
		i := strings.Index(s, word)
		if i < 0 {
			break
		}
		end := i + len(word)
		if (i > 0 && isNameChar(s[i-1])) || (end < len(s) && isNameChar(s[end])) {
			b.WriteString(s[:end])
		} else {
			b.WriteString(s[:i] + with)
		}
		s = s[end:]
	}
	b.WriteString(s)
	return b.String()
}

func isNameChar(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-' || c == '.' || c == '@'
}

// analyse reads the shape of the member's file.
func (m *member) analyse(refs References) {
	m.refsAt = -1
	wanted := sortedPaths(refs(m.project, nil))
	m.deps = len(wanted)
	tok := tokensOf(m.project)
	text := strings.ReplaceAll(string(m.raw), "\r\n", "\n")
	entries, ok := topLevel(m.raw)
	if !ok {
		m.text = tok.abstract(text)
		return
	}
	m.isJSON = true
	hasRefs := false
	for _, e := range entries {
		if e.key == referencesKey {
			hasRefs = true
			var list []any
			if json.Unmarshal(e.raw, &list) == nil && slices.Equal(sortedPaths(list), wanted) {
				m.refsAt = len(m.fields)
				continue
			}
		}
		var compact bytes.Buffer
		if json.Compact(&compact, e.raw) != nil {
			compact.Write(e.raw)
		}
		m.fields = append(m.fields, field{e.key, tok.abstract(compact.String())})
	}
	m.refsWild = !hasRefs && m.deps == 0
	if !hasRefs && m.deps > 0 {
		// The project has dependencies and the file lists none: no shape without
		// the key fits it, because deriving the references is what it needs.
		m.refsAt = refsLacking
	}
}

// sortedPaths lists the paths of reference entries, sorted.
func sortedPaths(entries []any) []string {
	var out []string
	for _, e := range entries {
		m, _ := e.(map[string]any)
		if p, ok := m["path"].(string); ok {
			out = append(out, p)
		}
	}
	slices.Sort(out)
	return out
}

// fieldsKey identifies the fields of a JSON member, without the references that
// follow the dependencies.
func (m *member) fieldsKey() string {
	var b strings.Builder
	for _, f := range m.fields {
		b.WriteString(f.key + "\x00" + f.value + "\n")
	}
	return b.String()
}

// value is the member's value for a key, in the abstract form; ok is false
// when the file has no such key. References that follow the project's
// dependencies are the sentinel.
func (m *member) value(key string) (string, bool) {
	if key == referencesKey && m.refsAt >= 0 {
		return sentinelRefs, true
	}
	for _, f := range m.fields {
		if f.key == key {
			return f.value, true
		}
	}
	return "", false
}
