package compose

// Rivals are the ticked choices that cannot stay when pick is ticked: the ones
// on the same single-choice axis of the same language, like radio buttons. Two
// choices share an axis only when both are Many. A language, a plugin and a
// choice with no axis have no rivals. The catalogue decides which axes take
// one choice: no axis name is known here.
//
// Parameters: ticked are the choices ticked now; pick is the one being ticked.
// Returns: the ticked choices to untick, in the order they were ticked.
func Rivals(ticked []Item, pick Item) []Item {
	if pick.Kind != KindChoice || pick.Axis == "" {
		return nil
	}
	var out []Item
	for _, t := range ticked {
		if t.ID != pick.ID && t.Kind == KindChoice && t.Axis == pick.Axis && t.Language == pick.Language && !(t.Many && pick.Many) {
			out = append(out, t)
		}
	}
	return out
}
