package palette

// unavailable is true for a shell command that cannot be used now.
func (i item) unavailable() bool { return i.entry != nil && i.entry.Unavailable != "" }

// arrange puts the commands that can be used first. With nothing typed it
// also leaves out the ones that cannot (outside a workspace that is most of
// them); searching finds them again, with the reason.
func arrange(found []item, hideUnavailable bool) []item {
	out := make([]item, 0, len(found))
	var blocked []item
	for _, it := range found {
		if it.unavailable() {
			blocked = append(blocked, it)
			continue
		}
		out = append(out, it)
	}
	if hideUnavailable {
		return out
	}
	return append(out, blocked...)
}

// hidden counts the commands left out of the list because they cannot be used now.
func (p Palette) hidden() int {
	n := 0
	for _, it := range p.items {
		if it.unavailable() {
			n++
		}
	}
	return n
}
