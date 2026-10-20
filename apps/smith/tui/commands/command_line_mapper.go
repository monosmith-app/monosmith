package commands

import "strings"

// Line is the command line the user would type for smith's arguments: an
// argument with a space or another character a shell would split or expand is
// put in single quotes, so a printed line can be pasted into a script as it is.
func Line(argv []string) string {
	parts := make([]string, 0, len(argv)+1)
	parts = append(parts, "smith")
	for _, a := range argv {
		parts = append(parts, quote(a))
	}
	return strings.Join(parts, " ")
}

// quote puts a in single quotes when a shell would treat it as more than one word.
func quote(a string) string {
	if a != "" && strings.IndexFunc(a, func(r rune) bool { return !safe(r) }) < 0 {
		return a
	}
	return "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
}

func safe(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		return true
	}
	return strings.ContainsRune("_@%+=:,./-", r)
}
