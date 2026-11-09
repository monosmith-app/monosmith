package siblingfiles

import (
	"bytes"
	"encoding/json"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// value turns an abstract JSON value back into the project's own and decodes it.
func (t tokens) value(abstract string) (any, error) {
	var v any
	dec := json.NewDecoder(bytes.NewReader([]byte(t.concrete(abstract))))
	dec.UseNumber()
	err := dec.Decode(&v)
	return v, err
}

// renderWhole is the content of the file for the project, in the form of the
// shape. A file that already has the form is its own content, byte for byte.
// ok is false when the form cannot be rendered.
func renderWhole(s *shape, m *member, refs References) (string, bool) {
	if m.exists && m.fits(s) {
		return string(m.raw), true
	}
	tok := tokensOf(m.project)
	rep := s.rep
	if !rep.isJSON {
		return tok.concrete(rep.text), true
	}
	doc, err := jsondoc.ParseDocument(rep.raw)
	if err != nil {
		return "", false
	}
	for _, f := range rep.fields {
		v, err := tok.value(f.value)
		if err != nil || doc.SetJSON(f.key, v) != nil {
			return "", false
		}
	}
	if rep.refsAt >= 0 {
		if m.deps == 0 {
			doc.Delete(referencesKey)
		} else if doc.SetJSON(referencesKey, refs(m.project, m.raw)) != nil {
			return "", false
		}
	}
	return string(doc.Bytes()), true
}

// renderKeys are the keys of the convention that the project's file owns, with
// their values for the project, and the keys that the file has with another
// value (or lacks, and is not owned): those stay the project's.
func renderKeys(majority map[string]string, m *member, refs References) (keys map[string]any, differing []string) {
	keys = map[string]any{}
	tok := tokensOf(m.project)
	for k, want := range majority {
		have, present := m.value(k)
		switch {
		case k == referencesKey && m.refsAt >= 0:
			keys[k] = refs(m.project, m.raw)
		case k == referencesKey && !present && m.deps == 0:
			// No dependency and no key: nothing to write.
		case k == referencesKey && !present && m.deps > 0:
			keys[k] = refs(m.project, m.raw)
		case k != referencesKey && present && have == want, k != referencesKey && !present && m.owned:
			if v, err := tok.value(want); err == nil {
				keys[k] = v
			}
		default:
			differing = append(differing, k)
		}
	}
	return keys, differing
}
