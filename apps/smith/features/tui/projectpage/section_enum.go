package projectpage

import (
	"strconv"

	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Section is a tab of the detail of a project.
type Section int

const (
	// Tasks is the table of the project's tasks and how they last ran.
	Tasks Section = iota
	// Dependencies is what the project uses and what uses it.
	Dependencies
	// Runs are the runs of this session that ran a task of the project.
	Runs
	// Settings is the editor of the project's settings in smith.toml.
	Settings
)

// Sections are the tabs of the detail, in the order they are drawn.
var Sections = []Section{Tasks, Dependencies, Runs, Settings}

// Title is the name of the tab. A narrow screen says "Deps" for Dependencies.
//
// Parameters: narrow is true on a narrow screen.
// Returns: the name.
func (s Section) Title(narrow bool) string {
	switch s {
	case Dependencies:
		if narrow {
			return words.Say("projectpage.section.deps")
		}
		return words.Say("projectpage.section.dependencies")
	case Runs:
		return words.Say("projectpage.section.runs")
	case Settings:
		return words.Say("projectpage.section.settings")
	}
	return words.Say("projectpage.section.tasks")
}

// Step is the tab dir places after this one (-1 is the one before), going
// round at the ends.
//
// Parameters: dir is the number of tabs to move.
// Returns: the tab.
func (s Section) Step(dir int) Section {
	n := len(Sections)
	return Sections[((int(s)+dir)%n+n)%n]
}

// sectionPrefix starts the ID of every tab.
const sectionPrefix = "x:section:"

// ID is the click and button ID of the tab: "x:section:" and its number.
//
// Returns: the ID.
func (s Section) ID() string { return sectionPrefix + strconv.Itoa(int(s)) }

// SectionOf is the tab an ID names.
//
// Parameters: id is a click or button ID.
// Returns: the tab, and false when id names no tab.
func SectionOf(id string) (Section, bool) {
	for _, s := range Sections {
		if s.ID() == id {
			return s, true
		}
	}
	return Tasks, false
}
