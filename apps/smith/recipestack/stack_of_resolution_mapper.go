package recipestack

import (
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// StackOf turns a resolved recipe into the stack a project is built with:
//
//   - use.language and use.scaffold are the stack's language and scaffold
//   - use.structure and use.deploy are its structure and deploy plugins
//   - use.toolchain lists toolchain plugins in the order given, then the
//     plugin of every other key (a slot such as lint or test) follows in the
//     key's alphabetical order; a plugin named twice is kept once
//
// A version constraint after an @ (eslint@^1.2) was checked when the recipe
// was validated and is left out of the names. A recipe that chooses no
// language cannot make a project.
func StackOf(res recipe.Resolution) (protocol.Stack, error) {
	st := protocol.Stack{Description: describe(res)}
	var explicit, slots []string
	for _, u := range res.Uses {
		names := pluginNames(u.Plugins)
		if len(names) == 0 {
			continue
		}
		switch u.Key {
		case "language":
			st.Language = names[0]
		case "scaffold":
			st.Scaffold = names[0]
		case "structure":
			st.Structure = names
		case "deploy":
			st.Deploy = names
		case "toolchain":
			explicit = names
		default:
			slots = append(slots, names...)
		}
	}
	if st.Language == "" {
		return protocol.Stack{}, fmt.Errorf("the recipe %s chooses no language: add use.language (or a when block that does) to create a project", res.Name)
	}
	for _, name := range append(explicit, slots...) {
		if !slices.Contains(st.Toolchain, name) {
			st.Toolchain = append(st.Toolchain, name)
		}
	}
	return st, nil
}

// pluginNames drops the version constraint after an @.
func pluginNames(specs []string) []string {
	names := make([]string, 0, len(specs))
	for _, spec := range specs {
		name, _, _ := strings.Cut(spec, "@")
		names = append(names, name)
	}
	return names
}

// describe is the stack's description in `smith stacks`: where it came from.
func describe(res recipe.Resolution) string {
	d := "From the recipe " + res.Name
	if res.Version != "" {
		d += " " + res.Version
	}
	if a := Answers(res); len(a) > 0 {
		var parts []string
		for _, q := range res.Asked {
			if q.Answer != "" {
				parts = append(parts, q.Name+"="+q.Answer)
			}
		}
		d += " (" + strings.Join(parts, ", ") + ")"
	}
	return d + "; written by smith"
}

// Answers are the choices as they were answered, by choice name.
func Answers(res recipe.Resolution) map[string]string {
	answers := map[string]string{}
	for _, q := range res.Asked {
		if q.Answer != "" {
			answers[q.Name] = q.Answer
		}
	}
	return answers
}
