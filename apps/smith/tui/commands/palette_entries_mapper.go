package commands

import (
	"fmt"
	"strings"
)

// The sections of the command palette, in the order they are listed.
const (
	SectionBuild        = "Build"
	SectionProjects     = "Projects"
	SectionDependencies = "Dependencies"
	SectionRelease      = "Release"
	SectionHealth       = "Health"
	SectionGoTo         = "Go to"
	SectionStart        = "Start"
	SectionHelp         = "Help"
)

// SectionOrder is the order of the sections in the palette.
var SectionOrder = []string{SectionStart, SectionBuild, SectionProjects, SectionDependencies, SectionRelease, SectionHealth, SectionGoTo, SectionHelp}

// PaletteEntry is a command as the palette lists it: one line for one intent,
// with the command it runs as a hint.
type PaletteEntry struct {
	ID, Section, Title, Hint, Shortcut, Unavailable string
}

// paletteSections says which commands the palette lists and under which
// section. A command that is not here (a key of a pane, a second way to the
// same thing) has no line, so one intent has one entry.
var paletteSections = map[string]string{
	"file.new": SectionStart, "file.adopt": SectionStart,

	"run.build": SectionBuild, "run.build-all": SectionBuild,
	"run.test": SectionBuild, "run.test-all": SectionBuild,
	"run.lint": SectionBuild, "run.lint-all": SectionBuild,
	"run.typecheck": SectionBuild, "run.typecheck-all": SectionBuild, "run.verify": SectionBuild, "run.affected": SectionBuild, "run.stop": SectionBuild,

	"file.add": SectionProjects, "proj.remove": SectionProjects, "proj.affected": SectionProjects, "proj.graph": SectionProjects,
	"projects.install-package": SectionProjects, "projects.remove-package": SectionProjects,

	"proj.install": SectionDependencies, "deps.update": SectionDependencies, "deps.align": SectionDependencies,

	"rel.preview": SectionRelease, "rel.notes": SectionRelease, "rel.release": SectionRelease, "ws.ci": SectionRelease,

	"ws.doctor": SectionHealth, "ws.fix": SectionHealth, "proj.upgrade": SectionHealth, "proj.plugins": SectionHealth, "ws.info": SectionHealth,
	"ws.undo": SectionHealth, "ws.report": SectionHealth, "ws.derive-check": SectionHealth, "ws.derive": SectionHealth, "ws.why": SectionHealth, "ws.agent-file": SectionHealth,

	"view.home": SectionGoTo, "view.projects": SectionGoTo, "view.run": SectionGoTo, "view.doctor": SectionGoTo,
	"view.deps": SectionGoTo, "view.release": SectionGoTo, "view.settings": SectionGoTo,

	"cli.guide": SectionHelp, "help.keys": SectionHelp, "view.mouse": SectionHelp, "file.quit": SectionHelp,
}

// PaletteEntries lists the commands the palette offers in a state, grouped by
// section: what cannot work here is left out, and what is only unavailable
// for now carries its reason. The titles of a run name its scope.
func PaletteEntries(s State) []PaletteEntry {
	var out []PaletteEntry
	for _, section := range SectionOrder {
		for _, c := range all {
			if paletteSections[c.ID] != section || !c.Visible(s) || (c.NeedsWorkspace && !s.InWorkspace) {
				continue
			}
			if c.Kind == RunTask && !c.All && s.RunProject() == "" && hasAllSibling(c) {
				// Off Projects the scoped command would only repeat "Build all N".
				continue
			}
			title, hint := paletteText(c, s)
			out = append(out, PaletteEntry{ID: c.ID, Section: section, Title: title, Hint: hint, Shortcut: c.Shortcut, Unavailable: c.Unavailable(s)})
		}
	}
	return out
}

// paletteText is the plain label of a command and the command line it runs,
// shown beside it in a muted colour.
func paletteText(c Command, s State) (title, hint string) {
	switch {
	case c.Kind == RunTask && c.All:
		if s.Projects > 0 {
			return fmt.Sprintf("%s all %d", c.verb(), s.Projects), Line(RunArgv(c.Task, ""))
		}
		return c.verb() + " all", Line(RunArgv(c.Task, ""))
	case c.Kind == RunTask:
		if s.RunProject() == "" {
			if s.Projects > 0 {
				return fmt.Sprintf("%s all %d", c.verb(), s.Projects), Line(RunArgv(c.Task, ""))
			}
			return c.verb() + " all", Line(RunArgv(c.Task, ""))
		}
		return c.verb() + " " + s.RunProject(), Line(RunArgv(c.Task, s.RunProject()))
	case c.Kind == ShowTab:
		return "Go to " + c.Title, c.Hint
	case len(c.Argv) > 0:
		return c.Title, Line(c.Argv)
	}
	return c.Title, c.Hint
}

// hasAllSibling is true when another command runs the same task on every project.
func hasAllSibling(c Command) bool {
	for _, o := range all {
		if o.Kind == RunTask && o.All && o.Task == c.Task {
			return true
		}
	}
	return false
}

// verb is the first word of a task's title: "Build", "Test"...
func (c Command) verb() string {
	verb, _, _ := strings.Cut(c.Title, " ")
	return verb
}
