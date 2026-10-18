package recipestack

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// NameOf is the name of the stack a recipe's answers make: the recipe's name
// and each answer, in the order the choices were asked, lowercased with
// anything but letters, digits, "_" and "-" turned into "-"
// (team-python-library). It is a valid bare key in smith.toml.
func NameOf(res recipe.Resolution) string {
	parts := []string{res.Name}
	for _, q := range res.Asked {
		if q.Answer != "" {
			parts = append(parts, q.Answer)
		}
	}
	return slug(strings.Join(parts, "-"))
}

func slug(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_':
			b.WriteRune(r)
			dash = false
		default:
			if !dash && b.Len() > 0 {
				b.WriteByte('-')
			}
			dash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// Choose settles the name a stack is declared under. A stack that already
// exists under the name with the same plugins is reused (reuse true, nothing
// to write): the same recipe and answers always make the same stack. One that
// exists and differs (a recipe changed since, or a published stack with the
// name) is kept as it is, and the new stack gets a short hash of its plugins
// after the name.
func Choose(name string, st protocol.Stack, existing func(name string) (protocol.Stack, bool)) (final string, reuse bool, err error) {
	if have, ok := existing(name); ok {
		if samePlugins(have, st) {
			return name, true, nil
		}
		name = fmt.Sprintf("%s-%s", name, fingerprint(st))
		if have, ok := existing(name); ok {
			if samePlugins(have, st) {
				return name, true, nil
			}
			return "", false, fmt.Errorf("a stack named %s already exists with other plugins", name)
		}
	}
	return name, false, nil
}

// samePlugins compares what builds a project, not the stack's name or text.
func samePlugins(a, b protocol.Stack) bool {
	a.Name, a.Description, b.Name, b.Description = "", "", "", ""
	return reflect.DeepEqual(normalized(a), normalized(b))
}

// normalized makes an empty list and a missing one equal.
func normalized(st protocol.Stack) protocol.Stack {
	for _, list := range []*[]string{&st.Toolchain, &st.Structure, &st.Deploy} {
		if len(*list) == 0 {
			*list = nil
		}
	}
	return st
}

// fingerprint is six hex digits of the stack's plugins.
func fingerprint(st protocol.Stack) string {
	st.Name, st.Description = "", ""
	raw, _ := json.Marshal(normalized(st))
	return fmt.Sprintf("%x", sha256.Sum256(raw))[:6]
}
