package workspacerecipe

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ErrNoProjects is returned when the workspace has no project to read a
// language from.
var ErrNoProjects = errors.New("the workspace has no projects, so there is no language to read: add a project first, or write the recipe with smith recipe new")

// levelNames are the levels that are not task slots: a plugin that provides
// a task of one of these names stays in the toolchain.
var levelNames = []string{"language", "scaffold", "toolchain", "structure", "deploy"}

// Languages lists the languages of the workspace's projects, the one most
// projects use first; languages with as many projects are in name order.
func Languages(f Facts) []string {
	counts := languageCounts(f)
	out := make([]string, 0, len(counts))
	for lang := range counts {
		out = append(out, lang)
	}
	sort.Slice(out, func(i, j int) bool {
		if counts[out[i]] != counts[out[j]] {
			return counts[out[i]] > counts[out[j]]
		}
		return out[i] < out[j]
	})
	return out
}

// ReadEach is the reading of every language of the workspace, in the order
// of Languages.
func ReadEach(f Facts) []Reading {
	var out []Reading
	for _, lang := range Languages(f) {
		if r, err := Read(f, lang); err == nil {
			out = append(out, r)
		}
	}
	return out
}

// Read is the recipe the workspace describes for language: empty language is
// the one most projects use.
//
// The language is fixed. The stack most of its projects use gives the
// scaffold, the toolchain, the structure and the deploy plugins. A toolchain
// plugin that alone provides one task (eslint: lint) fixes that task's slot,
// when doing so keeps the order the toolchain applies in. The layouts that
// smith.toml turns on and that check the language join the structure.
//
// It returns ErrNoProjects for a workspace without projects, and an error
// when no project uses language.
func Read(f Facts, language string) (Reading, error) {
	langs := Languages(f)
	if len(langs) == 0 {
		return Reading{}, ErrNoProjects
	}
	if language == "" {
		language = langs[0]
	}
	if !slices.Contains(langs, language) {
		return Reading{}, fmt.Errorf("no project of this workspace uses the language %q; the languages here are %s", language, strings.Join(langs, ", "))
	}
	r := Reading{Language: language, Projects: languageCounts(f)[language]}
	r.Fixed = append(r.Fixed, "language="+language)

	st, name := mainStack(f, language)
	r.StackName = name
	if st.Scaffold != "" {
		r.Fixed = append(r.Fixed, "scaffold="+st.Scaffold)
	}
	rest, slots := splitSlots(st.Toolchain, f.Provides)
	if len(rest) > 0 {
		r.Fixed = append(r.Fixed, "toolchain="+strings.Join(rest, ","))
	}
	for _, slot := range sortedKeys(slots) {
		r.Fixed = append(r.Fixed, slot+"="+slots[slot])
	}
	structure := append([]string(nil), st.Structure...)
	for _, l := range f.Layouts {
		if (len(l.Languages) == 0 || slices.Contains(l.Languages, language)) && !slices.Contains(structure, l.Name) {
			structure = append(structure, l.Name)
		}
	}
	if len(structure) > 0 {
		r.Fixed = append(r.Fixed, "structure="+strings.Join(structure, ","))
	}
	if len(st.Deploy) > 0 {
		r.Fixed = append(r.Fixed, "deploy="+strings.Join(st.Deploy, ","))
	}
	return r, nil
}

// languageCounts is the number of projects of each language.
func languageCounts(f Facts) map[string]int {
	counts := map[string]int{}
	for _, p := range f.Projects {
		if p.Language != "" {
			counts[p.Language]++
		}
	}
	return counts
}

// mainStack is the stack most projects of language use, and its name; a tie
// goes to the name that sorts first. Projects without a stack count as the
// stack "", which has no plugins.
func mainStack(f Facts, language string) (protocol.Stack, string) {
	counts := map[string]int{}
	stacks := map[string]protocol.Stack{}
	for _, p := range f.Projects {
		if p.Language != language {
			continue
		}
		name := p.StackName
		if name == protocol.NoStack {
			name = ""
		}
		counts[name]++
		if name != "" {
			stacks[name] = p.Stack
		}
	}
	best := ""
	for _, name := range sortedKeys(counts) {
		if counts[name] > counts[best] {
			best = name
		}
	}
	return stacks[best], best
}

// splitSlots takes out of a toolchain the plugins that fill one task slot:
// a plugin that provides exactly one task that no other plugin of the
// toolchain provides. A stack made from a recipe applies the toolchain first
// and then the slots by name, so the split is kept only when that order is
// the toolchain's own; otherwise every plugin stays in the toolchain.
func splitSlots(toolchain []string, provides map[string][]string) (rest []string, slots map[string]string) {
	count := map[string]int{}
	for _, p := range toolchain {
		for _, task := range provides[p] {
			count[task]++
		}
	}
	slots = map[string]string{}
	for _, p := range toolchain {
		tasks := provides[p]
		if len(tasks) == 1 && count[tasks[0]] == 1 && !slices.Contains(levelNames, tasks[0]) {
			slots[tasks[0]] = p
			continue
		}
		rest = append(rest, p)
	}
	rebuilt := append([]string(nil), rest...)
	for _, slot := range sortedKeys(slots) {
		rebuilt = append(rebuilt, slots[slot])
	}
	if !slices.Equal(rebuilt, toolchain) {
		return toolchain, nil
	}
	return rest, slots
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
