package selection

import "github.com/charmbracelet/x/ansi"

// snap widens the cells [from, to) of a screen row so that no wide character is
// cut in half: from moves back to the first cell of the character it is in, to
// moves on to the end of the character before it. A row that ends sooner is not
// padded.
func snap(line string, from, to int) (int, int) {
	pos, state := 0, byte(0)
	for rest := line; rest != ""; {
		_, w, n, next := ansi.DecodeSequence(rest, state, nil)
		rest, state = rest[n:], next
		if w <= 0 {
			continue
		}
		if from >= pos && from < pos+w {
			from = pos
		}
		if to-1 >= pos && to-1 < pos+w {
			to = pos + w
		}
		pos += w
	}
	return from, to
}
