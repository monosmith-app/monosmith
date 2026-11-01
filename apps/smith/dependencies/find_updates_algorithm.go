package dependencies

import (
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Levels are the groups of the up report, in order.
var Levels = []string{"patch", "minor", "major"}

// Update is a dependency with a newer release than some declarations allow.
type Update struct {
	Plugin string `json:"plugin"`
	Name   string `json:"name"`
	// Current is the highest version declared; Latest the registry's.
	Current string `json:"current"`
	Latest  string `json:"latest"`
	// Level is patch, minor or major: how far Latest is from Current.
	Level string `json:"level"`
	// Sites are the declarations up rewrites, each keeping its operator.
	Sites []protocol.Declaration `json:"sites"`
	// Preset is true for a package listed in [dependencies] presets: a team's
	// shared preset, bumped as one change in every project (ADR 0060).
	Preset bool `json:"preset,omitempty"`
}

// Group is the heading of the update in the up report: "preset" for a preset
// package, else its Level.
func (u Update) Group() string {
	if u.Preset {
		return "preset"
	}
	return u.Level
}

// MarkPresets sets Preset on the updates whose package is in presets and puts
// them first, keeping the order of the others.
func MarkPresets(updates []Update, presets []string) {
	for i := range updates {
		updates[i].Preset = slices.Contains(presets, updates[i].Name)
	}
	sort.SliceStable(updates, func(i, j int) bool { return updates[i].Preset && !updates[j].Preset })
}

// Projects names the declaring projects, "(root)" for the root manifest.
func (u Update) Projects() []string {
	var out []string
	for _, s := range u.Sites {
		out = append(out, where(s))
	}
	return out
}

// FindUpdates lists the dependencies whose latest release is newer than a
// declaration of them. Peers, the workspace's own packages and specs smith
// does not rewrite are left out.
func FindUpdates(plugin string, declared protocol.DeclaredResult, latest map[string]string, local map[string]bool) []Update {
	byName := map[string]*Update{}
	for _, d := range declared.Declarations {
		l, known := latest[d.Name]
		if d.Peer || !d.Rewritable || local[d.Name] || !known {
			continue
		}
		_, v, ok := splitSpec(d.Spec)
		if !ok || !newer(l, v) {
			continue
		}
		u := byName[d.Name]
		if u == nil {
			u = &Update{Plugin: plugin, Name: d.Name, Latest: l}
			byName[d.Name] = u
		}
		if u.Current == "" || newer(v, u.Current) {
			u.Current = v
		}
		u.Sites = append(u.Sites, d)
	}
	var out []Update
	for _, u := range byName {
		u.Level = level(u.Current, u.Latest)
		out = append(out, *u)
	}
	sort.Slice(out, func(i, j int) bool {
		if li, lj := levelIndex(out[i].Level), levelIndex(out[j].Level); li != lj {
			return li < lj
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// level is how far latest moves from current: major when the first part
// changes, minor when the second does, else patch.
func level(current, latest string) string {
	c, l := strings.Split(current, "."), strings.Split(latest, ".")
	switch {
	case c[0] != l[0]:
		return "major"
	case len(c) < 2 || len(l) < 2 || c[1] != l[1]:
		return "minor"
	}
	return "patch"
}

func levelIndex(l string) int {
	for i, x := range Levels {
		if x == l {
			return i
		}
	}
	return len(Levels)
}

// rewrites are the spec changes that apply updates: each declaration keeps
// its operator and moves to the latest version.
func rewrites(updates []Update) []protocol.Rewrite {
	var out []protocol.Rewrite
	for _, u := range updates {
		for _, s := range u.Sites {
			op, _, _ := splitSpec(s.Spec)
			out = append(out, protocol.Rewrite{Declaration: s, Spec: op + u.Latest})
		}
	}
	return out
}
