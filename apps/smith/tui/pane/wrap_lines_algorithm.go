package pane

// WrapRows wraps every line wider than w onto further lines (each continuation
// indented by indent cells) and moves the row numbers in rowOf, a map from a
// line to what it stands for, to where those lines now are. A line keeps its
// row on its first wrapped line.
func WrapRows[V any](lines []string, rowOf map[int]V, w, indent int) ([]string, map[int]V) {
	out := make([]string, 0, len(lines))
	moved := make(map[int]V, len(rowOf))
	for i, l := range lines {
		if v, ok := rowOf[i]; ok {
			moved[len(out)] = v
		}
		if Width(l) <= w {
			out = append(out, l)
			continue
		}
		out = append(out, Wrap(l, w, indent)...)
	}
	return out, moved
}
