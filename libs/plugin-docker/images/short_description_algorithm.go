package images

import "strings"

// ShortDescriptionLimit is the longest short description Docker Hub takes.
const ShortDescriptionLimit = 100

// FitShortDescription puts text on one line and, when it is longer than
// Docker Hub's limit, cuts it at a word boundary and ends it with an
// ellipsis, so the result never exceeds ShortDescriptionLimit characters. cut
// reports whether anything was lost.
func FitShortDescription(text string) (fitted string, cut bool) {
	text = strings.Join(strings.Fields(text), " ")
	runes := []rune(text)
	if len(runes) <= ShortDescriptionLimit {
		return text, false
	}
	head := string(runes[:ShortDescriptionLimit-1])
	if next := runes[ShortDescriptionLimit-1]; next != ' ' {
		if i := strings.LastIndex(head, " "); i > 0 {
			head = head[:i]
		}
	}
	return strings.TrimRight(head, " ,;:.-") + "…", true
}
