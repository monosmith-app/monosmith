package guide

import "strings"

import "github.com/monosmith-app/monosmith/apps/smith/scaffold"

// suggestName is the typed name put into the shape smith asks for ("My Shop"
// becomes "my-shop"), or "" when that gives nothing valid or changes nothing.
func suggestName(kind string) func(map[string][]string, string) string {
	return func(answers map[string][]string, value string) string {
		k := kind
		if k == "" {
			k = first(answers, "kind")
		}
		var b strings.Builder
		dash := false
		for _, r := range strings.ToLower(value) {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				if dash && b.Len() > 0 {
					b.WriteByte('-')
				}
				if b.Len() == 0 && r >= '0' && r <= '9' {
					continue
				}
				b.WriteRune(r)
				dash = false
			} else {
				dash = true
			}
		}
		out := b.String()
		if out == "" || out == value || scaffold.ValidateName(k, out) != nil {
			return ""
		}
		return out
	}
}

// namePlaceholder suggests a name that suits what is being added.
func namePlaceholder(answers map[string][]string) string { return exampleName(first(answers, "kind")) }

func exampleName(kind string) string {
	switch kind {
	case "lib":
		return "shared-code"
	case "package":
		return "my-package"
	case "workspace":
		return "my-studio"
	}
	return "web"
}
