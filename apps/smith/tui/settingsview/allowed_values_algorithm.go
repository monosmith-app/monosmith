package settingsview

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/configedit"
)

// allowed says in plain words what a setting accepts.
func allowed(f configedit.Field) string {
	switch f.Type {
	case "boolean":
		return "true or false (Enter flips it)"
	case "integer":
		if f.Minimum != nil {
			return fmt.Sprintf("a whole number, %d or more", *f.Minimum)
		}
		return "a whole number"
	case "array":
		if len(f.Enum) > 0 {
			return "any of " + strings.Join(f.Enum, ", ") + ", separated by spaces"
		}
		return "a list: words separated by spaces"
	}
	if f.MinLength > 0 {
		return fmt.Sprintf("text, at least %d characters", f.MinLength)
	}
	return "text"
}
