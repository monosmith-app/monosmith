package siblingfiles

import (
	"slices"
	"strconv"
)

// shape is a form that several members share. rep is one member that has it.
type shape struct {
	rep *member
	n   int
}

// key identifies the form of a member that is not wild.
func (m *member) shapeKey() string {
	if !m.isJSON {
		return "text\x00" + m.text
	}
	return "json\x00" + strconv.Itoa(m.refsAt) + "\x00" + m.fieldsKey()
}

// fits reports whether the member's file has the form of the shape: the same
// text, or the same fields in the same order with the references in the same
// place. A file with no references and a project with no dependency fits a
// shape with or without the key.
func (m *member) fits(s *shape) bool {
	r := s.rep
	if m.isJSON != r.isJSON {
		return false
	}
	if !m.isJSON {
		return m.text == r.text
	}
	if m.fieldsKey() != r.fieldsKey() {
		return false
	}
	return m.refsAt == r.refsAt || m.refsWild
}

// wholeShape finds the form that most members share. It returns nil unless
// that form is a convention: MinSiblings members and two thirds of all.
func wholeShape(members []*member) *shape {
	shapes := map[string]*shape{}
	var order []string
	for _, m := range members {
		if m.refsWild || m.refsAt == refsLacking {
			continue
		}
		k := m.shapeKey()
		if shapes[k] == nil {
			shapes[k] = &shape{rep: m}
			order = append(order, k)
		}
		shapes[k].n++
	}
	slices.Sort(order)
	for _, m := range members {
		if !m.refsWild {
			continue
		}
		var best *shape
		for _, k := range order {
			if s := shapes[k]; m.fits(s) && (best == nil || s.n > best.n) {
				best = s
			}
		}
		if best == nil {
			k := m.shapeKey()
			best = &shape{rep: m}
			shapes[k] = best
			order = append(order, k)
		}
		best.n++
	}
	var best *shape
	for _, k := range order {
		if s := shapes[k]; best == nil || s.n > best.n {
			best = s
		}
	}
	if best == nil || !isShared(best.n, len(members)) {
		return nil
	}
	return best
}

// keyMajorities finds, for each top-level key of the JSON members, the value
// that a convention shares, in the abstract form. A key whose values do not
// reach a convention is the project's own and is left out.
func keyMajorities(members []*member) map[string]string {
	var jsonMembers []*member
	keys := map[string]bool{}
	for _, m := range members {
		if !m.isJSON {
			continue
		}
		jsonMembers = append(jsonMembers, m)
		for _, f := range m.fields {
			keys[f.key] = true
		}
		if m.refsAt >= 0 {
			keys[referencesKey] = true
		}
	}
	out := map[string]string{}
	for k := range keys {
		counts := map[string]int{}
		for _, m := range jsonMembers {
			if v, ok := m.value(k); ok {
				counts[v]++
			} else if k == referencesKey && m.refsWild {
				counts[sentinelRefs]++
			}
		}
		for v, n := range counts {
			if isShared(n, len(jsonMembers)) {
				out[k] = v
			}
		}
	}
	return out
}
