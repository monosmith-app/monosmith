package depsview

import (
	"maps"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/hyperlink"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// DetailTitle is the title of the detail: the package under the cursor, or the
// one-version policy when no package is under it.
func (v Model) DetailTitle() string {
	if r, ok := v.Selected(); ok {
		return r.Name
	}
	return PolicyTitle()
}

// Detail tells about the package under the cursor: its versions and level,
// the projects that use it, how they declare it, its page on the registry, the
// one-version policy and the command line that updates it. With no package
// under the cursor it is the policy alone.
//
// Parameters: th draws, w is the width of the detail's text and root gives the
// folder of a project ("" when it is not known).
// Returns: the sections of the detail, in order.
func (v Model) Detail(th theme.Theme, w int, root func(project string) string) []detail.Section {
	policy := detail.Section{Heading: PolicyTitle(), Items: v.PolicyItems(th)}
	r, ok := v.Selected()
	if !ok {
		return []detail.Section{policy}
	}
	head := []string{r.Current + " → " + th.Bold(r.Latest) + "  " + LevelMark(th, r.Level)}
	if r.Level == "major" {
		head = append(head, pane.Wrap(th.Warning(words.Say("depspage.detail.major")), w, 0)...)
	}
	out := []detail.Section{{Items: head}}
	nameW := 0
	for _, p := range r.Projects {
		nameW = max(nameW, pane.Width(p))
	}
	var used []string
	for _, p := range r.Projects {
		line := pane.Cell(p, nameW+2)
		if root != nil {
			line += th.Muted(root(p))
		}
		used = append(used, line)
	}
	out = append(out, detail.Section{Heading: words.Say("depspage.detail.used_by"), Items: used, OneLine: true})
	if len(r.Kinds) > 0 {
		out = append(out, detail.Section{Heading: words.Say("depspage.detail.declared"), Items: []string{strings.Join(r.Kinds, ", ")}})
	}
	if line := OneVersionLine(th, r); line != "" {
		// The package's own verdict comes before the workspace's policy.
		policy.Items = append([]string{line}, policy.Items...)
	}
	out = append(out, detail.Section{Heading: words.Say("depspage.detail.registry"), Items: v.registryLines(th, r, w), OneLine: true}, policy)
	out = append(out, detail.Section{Heading: words.Say("depspage.detail.same_as"), Items: []string{th.Faint(words.Say("depspage.detail.command", "command", strings.Join(PackageArgv(r), " ")))}, OneLine: true})
	return out
}

// OneVersionLine says whether the projects that use a package declare it at
// one version: "✓ both use ^19.1.0" in the success colour, or "▲ web ^1.6.0 ·
// api ^1.5.0" in the warning colour when they differ.
//
// Parameters: th draws; r is the package.
// Returns: the line, "" when smith does not know how the package is declared.
func OneVersionLine(th theme.Theme, r Row) string {
	if len(r.Declared) == 0 {
		return ""
	}
	names := slices.Sorted(maps.Keys(r.Declared))
	specs := map[string]bool{}
	for _, n := range names {
		specs[r.Declared[n]] = true
	}
	if len(specs) == 1 {
		spec := r.Declared[names[0]]
		switch len(names) {
		case 1:
			return th.Success(th.Icons.Passed + " " + words.Say("depspage.one_version.one", "project", projectName(names[0]), "spec", spec))
		case 2:
			return th.Success(th.Icons.Passed + " " + words.Say("depspage.one_version.both", "spec", spec))
		}
		return th.Success(th.Icons.Passed + " " + words.Say("depspage.one_version.all", "count", len(names), "spec", spec))
	}
	parts := make([]string, 0, len(names))
	for _, n := range names {
		parts = append(parts, projectName(n)+" "+r.Declared[n])
	}
	return th.Warning(th.Icons.Warning + " " + strings.Join(parts, " · "))
}

// projectName is a project's name, or the words for the workspace root.
func projectName(name string) string {
	if name == "" {
		return words.Say("depspage.one_version.root")
	}
	return name
}

// registryLines say where the package is on its registry, as a hyperlink where
// the terminal draws one, and which keys open and copy the address. A package
// whose page cannot be known says so.
func (v Model) registryLines(th theme.Theme, r Row, w int) []string {
	if r.URL == "" {
		return []string{th.Muted(words.Say("depsview.registry.unknown", "name", r.Name))}
	}
	text := pane.MiddleTruncate(r.URL, max(w, 10))
	return []string{hyperlink.Link(r.URL, th.Info(text), v.Links), th.Muted(words.Say("depsview.registry.keys"))}
}

// PackageArgv is the command line that updates one package: `deps up <name>`,
// with --major for a major release, which the command takes only when asked.
//
// Parameters: r is the package.
// Returns: the arguments after smith.
func PackageArgv(r Row) []string {
	argv := []string{"deps", "up", r.Name}
	if r.Level == "major" {
		return append(argv, "--major")
	}
	return argv
}
