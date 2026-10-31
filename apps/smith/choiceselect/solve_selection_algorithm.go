package choiceselect

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Solve checks the named choices against the offers and returns the
// selection: the choices in the order they apply, with the default of every
// axis that a choice needs and the names leave out.
//
// A selection has one choice per axis, unless the choices on an axis all say
// Many; every Requires is in it, no Conflicts is in it, all choices are for one
// language, exactly one writes the project, and a choice that only adds to a
// plugin's project has that plugin in the selection. The first problem is the
// error, in one sentence that names the choices at fault and the fix.
func Solve(offers []Offer, names []string) (Selection, error) {
	index, err := Index(offers)
	if err != nil {
		return Selection{}, err
	}
	var chosen []Offer
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || slices.ContainsFunc(chosen, func(o Offer) bool { return o.Name == name }) {
			continue
		}
		o, ok := index[name]
		if !ok {
			return Selection{}, unknown(name, index)
		}
		chosen = append(chosen, o)
	}
	if len(chosen) == 0 {
		return Selection{}, errors.New("no choice is named: give at least the choice that writes the project, for example --with react (smith choices lists them)")
	}
	if err := checkSet(chosen, index); err != nil {
		return Selection{}, err
	}
	var added []string
	for i := 0; i < len(chosen); i++ {
		for _, axis := range chosen[i].Needs {
			if hasAxis(chosen, axis) {
				continue
			}
			fill, ok := defaultFor(offers, axis, chosen[i].Language)
			if !ok {
				return Selection{}, fmt.Errorf("%s needs a choice on the %s axis, and none is the default: add one of %s to --with", chosen[i].Name, axis, strings.Join(onAxis(offers, axis, chosen[i].Language), ", "))
			}
			chosen = append(chosen, fill)
			added = append(added, fill.Name)
		}
	}
	if len(added) > 0 {
		if err := checkSet(chosen, index); err != nil {
			return Selection{}, fmt.Errorf("%w (%s %s added because the axis is needed)", err, strings.Join(added, ", "), plural(len(added), "was", "were"))
		}
	}
	if err := checkRoles(chosen, offers); err != nil {
		return Selection{}, err
	}
	return Selection{Choices: order(chosen), Added: added}, nil
}

// Index maps each choice name to its offer. Two plugins that offer the same
// name are an error that names both.
func Index(offers []Offer) (map[string]Offer, error) {
	index := map[string]Offer{}
	for _, o := range offers {
		if have, dup := index[o.Name]; dup {
			return nil, fmt.Errorf("the plugins %s and %s both offer a choice named %s: disable one of them", have.Plugin, o.Plugin, o.Name)
		}
		index[o.Name] = o
	}
	return index, nil
}

// checkSet applies the rules that need only the chosen set: one language, one
// choice per axis, no conflict, every requirement met.
func checkSet(chosen []Offer, index map[string]Offer) error {
	for _, o := range chosen[1:] {
		if o.Language != chosen[0].Language {
			return fmt.Errorf("%s is for %s projects and %s for %s: keep choices of one language", chosen[0].Name, chosen[0].Language, o.Name, o.Language)
		}
	}
	for i, a := range chosen {
		for _, b := range chosen[i+1:] {
			if a.Axis == b.Axis && (!a.Many || !b.Many) {
				return fmt.Errorf("%s and %s are both on the %s axis: keep one", a.Name, b.Name, a.Axis)
			}
		}
	}
	for _, a := range chosen {
		for _, b := range chosen {
			if slices.Contains(a.Conflicts, b.Name) {
				return fmt.Errorf("%s conflicts with %s: drop one of them", a.Name, b.Name)
			}
		}
	}
	for _, a := range chosen {
		for _, need := range a.Requires {
			if slices.ContainsFunc(chosen, func(o Offer) bool { return o.Name == need }) {
				continue
			}
			if _, offered := index[need]; !offered {
				return fmt.Errorf("%s requires %s, which no enabled plugin offers: drop %s, or enable the plugin that offers %s", a.Name, need, a.Name, need)
			}
			return fmt.Errorf("%s requires %s, which is not in the selection: add %s to --with, or drop %s", a.Name, need, need, a.Name)
		}
	}
	return nil
}

// checkRoles applies the rules about who builds the project: one choice writes
// it, and a choice with no role has the plugin it adds to in the selection.
func checkRoles(chosen, offers []Offer) error {
	var scaffolds []Offer
	owners := map[string]bool{}
	for _, o := range chosen {
		if o.Plays(protocol.RoleScaffold) {
			scaffolds = append(scaffolds, o)
		}
		if o.Plays(protocol.RoleScaffold) || o.Plays(protocol.RoleToolchain) {
			owners[o.Plugin] = true
		}
	}
	switch {
	case len(scaffolds) == 0:
		return fmt.Errorf("no choice in %s writes the project: add one of %s", strings.Join(namesOf(chosen), ", "), strings.Join(withRole(offers, protocol.RoleScaffold, chosen[0].Language), ", "))
	case len(scaffolds) > 1:
		return fmt.Errorf("%s and %s both write the project: keep one", scaffolds[0].Name, scaffolds[1].Name)
	}
	for _, o := range chosen {
		if len(o.Roles) > 0 || owners[o.Plugin] {
			continue
		}
		var users []string
		for _, other := range offers {
			if other.Plugin == o.Plugin && other.Language == o.Language && len(other.Roles) > 0 {
				users = append(users, other.Name)
			}
		}
		sort.Strings(users)
		return fmt.Errorf("%s is provided by the %s plugin, which no other choice here uses: add one of %s to --with", o.Name, o.Plugin, strings.Join(users, ", "))
	}
	return nil
}

// order puts the choice that writes the project first and every choice after
// those it requires, otherwise in the order given.
func order(chosen []Offer) []Offer {
	byName := map[string]Offer{}
	for _, o := range chosen {
		byName[o.Name] = o
	}
	var out []Offer
	done := map[string]bool{}
	var visit func(o Offer)
	visit = func(o Offer) {
		if done[o.Name] {
			return
		}
		done[o.Name] = true
		for _, r := range o.Requires {
			if dep, ok := byName[r]; ok {
				visit(dep)
			}
		}
		out = append(out, o)
	}
	for _, o := range chosen {
		if o.Plays(protocol.RoleScaffold) {
			visit(o)
		}
	}
	for _, o := range chosen {
		visit(o)
	}
	return out
}

func hasAxis(chosen []Offer, axis string) bool {
	return slices.ContainsFunc(chosen, func(o Offer) bool { return o.Axis == axis })
}

func defaultFor(offers []Offer, axis, language string) (Offer, bool) {
	for _, o := range offers {
		if o.Axis == axis && o.Language == language && o.Default {
			return o, true
		}
	}
	return Offer{}, false
}

func onAxis(offers []Offer, axis, language string) []string {
	var out []string
	for _, o := range offers {
		if o.Axis == axis && o.Language == language {
			out = append(out, o.Name)
		}
	}
	sort.Strings(out)
	return out
}

func withRole(offers []Offer, role protocol.Role, language string) []string {
	var out []string
	for _, o := range offers {
		if o.Language == language && o.Plays(role) && o.Axis != "" {
			out = append(out, o.Name)
		}
	}
	sort.Strings(out)
	return out
}

func namesOf(offers []Offer) []string {
	out := make([]string, len(offers))
	for i, o := range offers {
		out[i] = o.Name
	}
	return out
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// unknown is the error for a name no plugin offers, with the closest names.
func unknown(name string, index map[string]Offer) error {
	var near []string
	for other := range index {
		if distance(name, other) <= 2 {
			near = append(near, other)
		}
	}
	sort.Strings(near)
	if len(near) > 0 {
		return fmt.Errorf("no choice is named %q: did you mean %s? (smith choices lists them)", name, strings.Join(near, " or "))
	}
	return fmt.Errorf("no choice is named %q (smith choices lists them)", name)
}

// distance is the edit distance between two names.
func distance(a, b string) int {
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(b)]
}
