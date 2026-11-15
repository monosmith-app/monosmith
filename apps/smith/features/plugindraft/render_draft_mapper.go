package plugindraft

import (
	"fmt"
	"strings"
)

// Render writes the draft as a TOML file a person edits, with a comment where
// each value came from. The file is input for the plugin author (or an
// assistant): smith reads it nowhere.
//
// Parameters: lang is the language name; sample is the folder inspected.
func Render(lang, sample string, d Draft) string {
	var b strings.Builder
	q := func(s string) string { return fmt.Sprintf("%q", s) }
	fmt.Fprintf(&b, "# Draft of a language plugin for %q, made by `smith plugin draft` from %s.\n", lang, sample)
	b.WriteString("# It lists what the sample shows. Check every line: a guess is marked \"guess\".\n\n")
	fmt.Fprintf(&b, "[plugin]\nname = %s\nkind = \"language\"\nlanguages = [%s]\n", q("lang-"+lang), q(lang))
	var files []string
	for _, m := range d.Manifests {
		files = append(files, q(m.File))
	}
	fmt.Fprintf(&b, "manifests = [%s] # guess: files that name a project and give it a version\n", strings.Join(files, ", "))
	for _, m := range d.Manifests {
		fmt.Fprintf(&b, "\n[[manifest]]\nfile = %s\nname_line = %s\nversion_line = %s # version/write rewrites this line\n", q(m.File), q(m.NameLine), q(m.VersionLine))
	}
	for _, t := range d.Tasks {
		var parts []string
		for _, c := range t.Command {
			parts = append(parts, q(c))
		}
		fmt.Fprintf(&b, "\n[tasks.%s] # from %s\ncommand = [%s]\n", t.Name, t.Source, strings.Join(parts, ", "))
	}
	if len(d.Publish) > 0 {
		b.WriteString("\n[publish] # lines that look like a publish step; pick one\ncandidates = [\n")
		for _, p := range d.Publish {
			fmt.Fprintf(&b, "  %s,\n", q(p))
		}
		b.WriteString("]\n")
	}
	return b.String()
}

// Todo is what a plugin author still writes, as lines for the console.
func Todo(d Draft) []string {
	out := []string{"discover: turn each manifest into a project (its kind, its imports, its exports)"}
	if len(d.Manifests) == 0 {
		out = append([]string{"manifests: none was found; name the file that marks a project"}, out...)
	}
	out = append(out,
		"version/write: rewrite the version line in the manifest",
		"generate: the starter files of a new project",
	)
	if len(d.Publish) == 0 {
		out = append(out, "publish and published: no publish step was found in the sample")
	} else {
		out = append(out, "publish and published: choose the publish line, and how to ask the registry what it has")
	}
	return append(out, "ci: what a pipeline must install and cache")
}
