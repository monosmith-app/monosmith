package dependencies

import (
	"strings"
)

// describeMajors lists major updates as "name 1.2.0 -> 3.0.0".
func describeMajors(us []Update) string {
	var parts []string
	for _, u := range us {
		parts = append(parts, u.Name+" "+u.Current+" -> "+u.Latest)
	}
	return strings.Join(parts, ", ")
}
