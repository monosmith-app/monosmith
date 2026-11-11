package commands

import (
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// The sections of the command palette, in the order they are listed.
var (
	SectionBuild        = words.Say("commands.section.build")
	SectionProjects     = words.Say("commands.section.projects")
	SectionDependencies = words.Say("commands.section.dependencies")
	SectionRelease      = words.Say("commands.section.release")
	SectionHealth       = words.Say("commands.section.health")
	SectionGoTo         = words.Say("commands.section.goto")
	SectionStart        = words.Say("commands.section.start")
	SectionHelp         = words.Say("commands.section.help")
)

// SectionOrder is the order of the sections in the palette.
var SectionOrder = []string{SectionStart, SectionBuild, SectionProjects, SectionDependencies, SectionRelease, SectionHealth, SectionGoTo, SectionHelp}

// PaletteEntry is a command as the palette lists it: one line for one intent,
// with the command it runs as a hint. Keywords are the synonyms the catalogue
// holds for the command: the search finds them, the list never shows them.
type PaletteEntry struct {
	ID, Section, Title, Hint, Shortcut, Keywords, Unavailable string
}

// keywordsKey is the catalogue key prefix of the synonyms of a command.
const keywordsKey = "palette.keywords."

// keywordsOf is the synonyms of a command from the catalogue, or "" when it has
// none. A key the catalogue lacks gives the key itself, which means no text.
func keywordsOf(id string) string {
	key := keywordsKey + id
	if text := words.Say(key); text != key {
		return text
	}
	return ""
}

// paletteSections says which commands the palette lists and under which
// section. A command that is not here (a key of a pane, a second way to the
// same thing) has no line, so one intent has one entry.
var paletteSections = map[string]string{
	"file.new": SectionStart, "file.adopt": SectionStart, "file.compose": SectionStart, "ws.commit": SectionHealth, "ws.tools": SectionHealth,

	"run.build": SectionBuild, "run.build-all": SectionBuild, "run.build-release": SectionBuild, "run.build-release-all": SectionBuild,
	"run.test": SectionBuild, "run.test-all": SectionBuild,
	"run.lint": SectionBuild, "run.lint-all": SectionBuild,
	"run.typecheck": SectionBuild, "run.typecheck-all": SectionBuild, "run.verify": SectionBuild, "run.affected": SectionBuild, "run.stop": SectionBuild,

	"file.add": SectionProjects, "proj.swap": SectionProjects, "proj.remove": SectionProjects, "proj.affected": SectionProjects, "proj.graph": SectionProjects,
	"projects.install-package": SectionProjects, "projects.remove-package": SectionProjects,

	"proj.install": SectionDependencies, "deps.update": SectionDependencies, "deps.align": SectionDependencies, "deps.one-version": SectionDependencies,

	"rel.preview": SectionRelease, "rel.notes": SectionRelease, "rel.release": SectionRelease, "rel.forced": SectionRelease, "ws.ci": SectionRelease,

	"ws.doctor": SectionHealth, "ws.fix": SectionHealth, "proj.upgrade": SectionHealth, "proj.plugins": SectionHealth, "proj.plugin-search": SectionHealth, "ws.info": SectionHealth,
	"ws.undo": SectionHealth, "ws.report": SectionHealth, "ws.completion": SectionHealth, "ws.derive-check": SectionHealth, "ws.derive": SectionHealth, "ws.derive-preview": SectionHealth, "ws.why": SectionHealth, "ws.divergences": SectionHealth, "ws.try": SectionHealth, "ws.bootstrap": SectionHealth, "ws.agent-file": SectionHealth,
	"ws.cache-setup": SectionHealth, "ws.cache-status": SectionHealth, "ws.cache-verify": SectionHealth, "ws.cache-repair": SectionHealth, "ws.cache-prune": SectionHealth, "ws.cache-clear": SectionHealth,
	"ws.copies": SectionHealth, "ws.tools-list": SectionHealth, "ws.why-miss": SectionHealth, "ws.mcp-install": SectionHealth,
	"ws.env-list": SectionProjects, "ws.env-add": SectionProjects, "ws.env-remove": SectionProjects, "ws.pipeline": SectionRelease, "ws.tauri-icon": SectionProjects, "ws.docker-description": SectionRelease,

	"view.home": SectionGoTo, "view.projects": SectionGoTo, "view.doctor": SectionGoTo,
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
			out = append(out, PaletteEntry{ID: c.ID, Section: section, Title: title, Hint: hint, Shortcut: c.Shortcut, Keywords: keywordsOf(c.ID), Unavailable: c.Unavailable(s)})
		}
		if section == SectionBuild && s.Project != "" && s.InWorkspace {
			// The selected project's own tasks (dev, storybook…) have no key of
			// their own: the palette is their keyboard path.
			busy := ""
			if s.Running {
				busy = words.Say("commands.reason.running", "shortcut", ShortcutOf("run.stop"))
			}
			for _, task := range s.OtherTasks {
				out = append(out, PaletteEntry{ID: "x:run:" + task, Section: section, Title: words.Say("commands.palette.task", "task", task, "project", s.Project), Hint: Line(RunArgv(task, s.Project)), Keywords: task, Unavailable: busy})
			}
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
			return words.Say("commands.palette.all.count", "verb", c.verb(), "count", s.countFor(c.Task)), Line(c.TaskArgv(""))
		}
		return words.Say("commands.palette.all", "verb", c.verb()), Line(c.TaskArgv(""))
	case c.Kind == RunTask:
		if s.RunProject() == "" {
			if s.Projects > 0 {
				return words.Say("commands.palette.all.count", "verb", c.verb(), "count", s.countFor(c.Task)), Line(c.TaskArgv(""))
			}
			return words.Say("commands.palette.all", "verb", c.verb()), Line(c.TaskArgv(""))
		}
		return c.verb() + " " + s.RunProject(), Line(c.TaskArgv(s.RunProject()))
	case c.Kind == ShowTab:
		return words.Say("commands.palette.goto", "title", c.Title), c.Hint
	case len(c.Argv) > 0:
		return c.Title, Line(c.Argv)
	case s.On[c.ID] && c.OnTitle != "":
		return c.OnTitle, c.Hint
	}
	return c.Title, c.Hint
}

// hasAllSibling is true when another command runs the same task on every project.
func hasAllSibling(c Command) bool {
	for _, o := range all {
		if o.Kind == RunTask && o.All && o.Task == c.Task && o.Release == c.Release {
			return true
		}
	}
	return false
}

// verb is the name of a task as the palette shows it with its scope: "Build",
// "Lint"... The title of the command for one project, so the "all"
// command borrows it.
func (c Command) verb() string {
	if c.All {
		for _, o := range all {
			if o.Kind == RunTask && !o.All && o.Task == c.Task && o.Release == c.Release {
				return o.Title
			}
		}
	}
	return c.Title
}
