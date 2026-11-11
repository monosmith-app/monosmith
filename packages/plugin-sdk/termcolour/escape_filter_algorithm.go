package termcolour

import "strings"

const esc = 0x1b

// KeepSGR removes every terminal escape sequence from s except SGR colour
// codes (ESC [ ... m): cursor movement, line erasing, window titles,
// hyperlinks and the like are dropped, so a forced-colour tool cannot draw a
// progress bar or a spinner into a log.
//
// Parameters: s is one line of tool output.
// Returns: s with only its SGR codes left.
func KeepSGR(s string) string { return filter(s, true) }

// Strip removes every terminal escape sequence from s, colour included.
//
// Parameters: s is text that may carry escape sequences.
// Returns: the plain text.
func Strip(s string) string { return filter(s, false) }

// HasSGR reports whether s carries at least one SGR colour code.
//
// Parameters: s is one line of text.
// Returns: true when the line is already coloured.
func HasSGR(s string) bool {
	for i := 0; i < len(s); {
		j := strings.IndexByte(s[i:], esc)
		if j < 0 {
			return false
		}
		i += j
		n, sgr := sequence(s[i:])
		if sgr {
			return true
		}
		i += n
	}
	return false
}

func filter(s string, keepSGR bool) string {
	if strings.IndexByte(s, esc) < 0 {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] != esc {
			b.WriteByte(s[i])
			i++
			continue
		}
		n, sgr := sequence(s[i:])
		if sgr && keepSGR {
			b.WriteString(s[i : i+n])
		}
		i += n
	}
	return b.String()
}

// sequence measures the escape sequence at the start of s (s[0] is ESC) and
// reports whether it is an SGR code. An unfinished sequence runs to the end.
func sequence(s string) (n int, sgr bool) {
	if len(s) < 2 {
		return len(s), false
	}
	switch s[1] {
	case '[': // CSI: parameters 0x30-0x3F, intermediates 0x20-0x2F, final 0x40-0x7E.
		i := 2
		params := true
		for i < len(s) && s[i] >= 0x30 && s[i] <= 0x3f {
			if (s[i] < '0' || s[i] > '9') && s[i] != ';' && s[i] != ':' {
				params = false
			}
			i++
		}
		inter := i
		for i < len(s) && s[i] >= 0x20 && s[i] <= 0x2f {
			i++
		}
		if i < len(s) && s[i] >= 0x40 && s[i] <= 0x7e {
			return i + 1, s[i] == 'm' && params && inter == i
		}
		return i, false
	case ']', 'P', '_', '^', 'X': // OSC, DCS, APC, PM, SOS: up to BEL or ESC \.
		for i := 2; i < len(s); i++ {
			if s[i] == 0x07 {
				return i + 1, false
			}
			if s[i] == esc && i+1 < len(s) && s[i+1] == '\\' {
				return i + 2, false
			}
		}
		return len(s), false
	}
	// Any other escape: intermediates, then one final byte (ESC 7, ESC ( B).
	i := 1
	for i < len(s) && s[i] >= 0x20 && s[i] <= 0x2f {
		i++
	}
	if i < len(s) {
		i++
	}
	return i, false
}
