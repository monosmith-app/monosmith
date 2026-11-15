package convertplan

import (
	"regexp"
	"strings"
)

// nxReactTypings is the prefix of the declaration files @nx/react's
// generator lists in a tsconfig's compilerOptions.types (cssmodule.d.ts,
// image.d.ts), which go with @nx/react.
const nxReactTypings = "@nx/react/typings/"

var (
	typesArray = regexp.MustCompile(`("types"\s*:\s*\[)([^\]]*)(\])`)
	typesItem  = regexp.MustCompile(`"([^"]*)"`)
	itemIndent = regexp.MustCompile(`\n([ \t]*)"`)
	endIndent  = regexp.MustCompile(`\n([ \t]*)$`)
)

// RewriteNxReactTypings replaces the @nx/react typings in a tsconfig's
// types with vite/client, which declares the same CSS-module and image
// imports for a Vite project. Only the types array changes, keeping the
// rest of the file (comments included) as it is. ok is false when the file
// lists none of them.
func RewriteNxReactTypings(text string) (string, bool) {
	m := typesArray.FindStringSubmatchIndex(text)
	if m == nil {
		return text, false
	}
	body := text[m[4]:m[5]]
	var kept []string
	found, vite := false, false
	for _, item := range typesItem.FindAllStringSubmatch(body, -1) {
		switch {
		case strings.HasPrefix(item[1], nxReactTypings):
			found = true
		default:
			vite = vite || item[1] == "vite/client"
			kept = append(kept, item[1])
		}
	}
	if !found {
		return text, false
	}
	if !vite {
		kept = append(kept, "vite/client")
	}
	if !strings.Contains(body, "\n") {
		return text[:m[4]] + `"` + strings.Join(kept, `", "`) + `"` + text[m[5]:], true
	}
	indent, end := "      ", "    "
	if i := itemIndent.FindStringSubmatch(body); i != nil {
		indent = i[1]
	}
	if e := endIndent.FindStringSubmatch(body); e != nil {
		end = e[1]
	}
	items := make([]string, len(kept))
	for i, k := range kept {
		items[i] = "\n" + indent + `"` + k + `"`
	}
	return text[:m[4]] + strings.Join(items, ",") + "\n" + end + text[m[5]:], true
}
