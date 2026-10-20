package pipeline

import "regexp"

// unsafeUse is an unsafe block, function, impl, trait or extern block.
var unsafeUse = regexp.MustCompile(`\bunsafe\s*(\{|fn\b|impl\b|trait\b|extern\b)`)

// CountUnsafe is the number of unsafe blocks, functions, impls and traits in
// Rust source, ignoring comments and string literals.
func CountUnsafe(src string) int {
	return len(unsafeUse.FindAllStringIndex(blankCommentsAndStrings(src), -1))
}

// blankCommentsAndStrings replaces comments and string contents with spaces,
// so a word in them never counts. Raw strings are read as ordinary ones.
func blankCommentsAndStrings(src string) string {
	b := []byte(src)
	blank := func(i int) {
		if b[i] != '\n' {
			b[i] = ' '
		}
	}
	for i := 0; i < len(b); {
		switch {
		case b[i] == '/' && i+1 < len(b) && b[i+1] == '/':
			for ; i < len(b) && b[i] != '\n'; i++ {
				blank(i)
			}
		case b[i] == '/' && i+1 < len(b) && b[i+1] == '*':
			depth := 0
			for i < len(b) {
				switch {
				case b[i] == '/' && i+1 < len(b) && b[i+1] == '*':
					depth++
					blank(i)
					blank(i + 1)
					i += 2
				case b[i] == '*' && i+1 < len(b) && b[i+1] == '/':
					depth--
					blank(i)
					blank(i + 1)
					i += 2
				default:
					blank(i)
					i++
				}
				if depth == 0 {
					break
				}
			}
		case b[i] == '\'' && i+2 < len(b) && b[i+2] == '\'':
			i += 3 // a char literal such as '"'
		case b[i] == '"':
			blank(i)
			i++
			for i < len(b) && b[i] != '"' {
				if b[i] == '\\' && i+1 < len(b) {
					blank(i)
					i++
				}
				blank(i)
				i++
			}
			if i < len(b) {
				blank(i)
			}
			i++
		default:
			i++
		}
	}
	return string(b)
}
