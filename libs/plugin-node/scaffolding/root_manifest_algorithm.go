package scaffolding

import (
	"bytes"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/lineendings"
)

// manifestMember is one member of a JSON object, its value left as written.
type manifestMember struct {
	key   string
	value json.RawMessage
}

// completeRootManifest adds to a root package.json the workspaces,
// devDependencies and overrides it lacks and says whether anything was added. Everything
// already there stays as written: the order of its keys, its scripts and the
// characters in them (&& is not escaped), so the edit shows in a diff as the
// lines it added. An empty raw starts a new manifest.
func completeRootManifest(raw []byte, workspaceName string, devDependencies, overrides map[string]string) ([]byte, bool, error) {
	changed := len(bytes.TrimSpace(raw)) == 0
	var members []manifestMember
	if changed {
		members = []manifestMember{
			{"name", quoted(strings.ToLower(workspaceName) + "-workspace")},
			{"private", json.RawMessage("true")},
		}
	} else {
		var err error
		if members, err = readManifestMembers(raw); err != nil {
			return nil, false, err
		}
	}
	if indexOfMember(members, "workspaces") < 0 {
		members = append(members, manifestMember{"workspaces", json.RawMessage(`["apps/*", "libs/*", "packages/*"]`)})
		changed = true
	}
	for _, field := range []struct {
		key    string
		wanted map[string]string
	}{{"devDependencies", devDependencies}, {"overrides", overrides}} {
		if len(field.wanted) == 0 {
			continue
		}
		value, added, err := withMissingEntries(members, field.key, field.wanted)
		if err != nil {
			return nil, false, err
		}
		if !added {
			continue
		}
		changed = true
		if i := indexOfMember(members, field.key); i >= 0 {
			members[i].value = value
		} else {
			members = append(members, manifestMember{field.key, value})
		}
	}
	if !changed {
		return raw, false, nil
	}
	out, err := renderManifestMembers(members)
	// A CRLF manifest stays CRLF.
	return []byte(lineendings.Match(string(raw), string(out))), true, err
}

// withMissingEntries is the object under key with the missing entries added,
// and whether any was. An entry already there is kept, whatever its value.
// Its keys stay sorted when they were.
func withMissingEntries(members []manifestMember, key string, wanted map[string]string) (json.RawMessage, bool, error) {
	var existing []manifestMember
	if i := indexOfMember(members, key); i >= 0 {
		var err error
		if existing, err = readManifestMembers(members[i].value); err != nil {
			return nil, false, err
		}
	}
	sorted := sort.SliceIsSorted(existing, func(a, b int) bool { return existing[a].key < existing[b].key })
	added := false
	names := make([]string, 0, len(wanted))
	for name := range wanted {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if indexOfMember(existing, name) < 0 {
			existing = append(existing, manifestMember{name, quoted(wanted[name])})
			added = true
		}
	}
	if sorted {
		sort.SliceStable(existing, func(a, b int) bool { return existing[a].key < existing[b].key })
	}
	var b bytes.Buffer
	b.WriteByte('{')
	for i, m := range existing {
		if i > 0 {
			b.WriteByte(',')
		}
		b.Write(quoted(m.key))
		b.WriteByte(':')
		b.Write(m.value)
	}
	b.WriteByte('}')
	return b.Bytes(), added, nil
}

// readManifestMembers reads a JSON object's members in the order written.
func readManifestMembers(raw []byte) ([]manifestMember, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	if open, err := dec.Token(); err != nil {
		return nil, err
	} else if d, ok := open.(json.Delim); !ok || d != '{' {
		return nil, errors.New("not a JSON object")
	}
	var members []manifestMember
	for dec.More() {
		key, err := dec.Token()
		if err != nil {
			return nil, err
		}
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return nil, err
		}
		members = append(members, manifestMember{key.(string), value})
	}
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	return members, nil
}

// renderManifestMembers writes the members as an object indented by two
// spaces, with a final newline.
func renderManifestMembers(members []manifestMember) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("{\n")
	for i, m := range members {
		b.WriteString("  ")
		b.Write(quoted(m.key))
		b.WriteString(": ")
		if err := json.Indent(&b, m.value, "  ", "  "); err != nil {
			return nil, err
		}
		if i < len(members)-1 {
			b.WriteByte(',')
		}
		b.WriteByte('\n')
	}
	b.WriteString("}\n")
	return b.Bytes(), nil
}

func indexOfMember(members []manifestMember, key string) int {
	for i, m := range members {
		if m.key == key {
			return i
		}
	}
	return -1
}

// quoted is a JSON string, with <, > and & as they are.
func quoted(s string) json.RawMessage {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	return bytes.TrimRight(b.Bytes(), "\n")
}
