package guide

// ToggleOption flips the mark of option i in a multiple choice and keeps the
// options of a single-choice axis exclusive, like radio buttons: when a tick
// turns on an option, every other marked option of the same Group is turned
// off, unless both options allow several (Option.Many).
//
// Parameters: marked are the marks by option number; options are the options;
// i is the option flipped.
// Returns: a new set of marks (the input is not changed) and, when ticking i
// unticked other options, a one-line note that names them.
func ToggleOption(marked map[int]bool, options []Option, i int) (map[int]bool, string) {
	out := make(map[int]bool, len(marked)+1)
	for k, v := range marked {
		out[k] = v
	}
	if i < 0 || i >= len(options) {
		return out, ""
	}
	out[i] = !out[i]
	if !out[i] || options[i].Group == "" {
		return out, ""
	}
	var dropped []string
	for j, o := range options {
		if j == i || !out[j] || o.Group != options[i].Group || (o.Many && options[i].Many) {
			continue
		}
		out[j] = false
		dropped = append(dropped, o.Value)
	}
	if len(dropped) == 0 {
		return out, ""
	}
	return out, options[i].Value + " replaces " + joinWords(dropped) + ": only one " + options[i].Group + " fits a project. Untick " + options[i].Value + " to get it back."
}

func joinWords(words []string) string {
	switch len(words) {
	case 1:
		return words[0]
	case 2:
		return words[0] + " and " + words[1]
	}
	return words[0] + ", " + joinWords(words[1:])
}
