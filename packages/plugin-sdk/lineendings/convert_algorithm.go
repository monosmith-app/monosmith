package lineendings

import "strings"

// Convert returns text with every line ending replaced by CRLF when crlf is
// true, else by LF. A mix of endings in text becomes uniform.
func Convert(text string, crlf bool) string {
	if !strings.Contains(text, "\n") {
		return text
	}
	lf := strings.ReplaceAll(text, "\r\n", "\n")
	if crlf {
		return strings.ReplaceAll(lf, "\n", "\r\n")
	}
	return lf
}

// IsCRLF reports whether text mostly ends its lines with CRLF; known is false
// when text has no line ending to judge by.
func IsCRLF(text string) (crlf, known bool) {
	total := strings.Count(text, "\n")
	if total == 0 {
		return false, false
	}
	return strings.Count(text, "\r\n")*2 > total, true
}

// Match returns content with the line ending that existing uses. It returns
// content unchanged when existing has no line ending to follow.
func Match(existing, content string) string {
	crlf, known := IsCRLF(existing)
	if !known {
		return content
	}
	return Convert(content, crlf)
}
