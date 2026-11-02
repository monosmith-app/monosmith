package protocol

import "slices"

// Choice is one option on an axis of a project's make-up: a framework, a test
// runner, a style library. A plugin publishes the choices it owns in
// InitializeResult.Choices; the host checks a selection of them (smith add
// --with react,vitest,bootstrap) and turns it into a stack. What a choice
// writes (dependencies, tasks, files) stays with the plugin: the host never
// sees it.
type Choice struct {
	// Axis is the question the choice answers: "framework", "test", "style",
	// "lint". The plugins agree on the name; the host only compares it.
	Axis string `json:"axis"`
	// Name identifies the choice in a selection and is unique among the
	// choices of all loaded plugins.
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// Language is the language plugin the choice applies to ("node").
	Language string `json:"language"`
	// Roles are the roles the choice's plugin plays in a stack built from it:
	// scaffold (it writes the project) and/or toolchain (it joins the
	// toolchain). Empty means it only adds to what a plugin of the selection
	// already writes, as a style does.
	Roles []Role `json:"roles,omitempty"`
	// Needs are axes that the selection must have a choice on. An axis the
	// person leaves out is filled with the default of the axis.
	Needs []string `json:"needs,omitempty"`
	// Default marks the choice the host adds when a selection needs its axis
	// and has no choice on it.
	Default bool `json:"default,omitempty"`
	// Suggested marks the choice an asker pre-selects on its axis. The solver
	// never adds it by itself: an axis that a choice needs and that has no
	// Default is asked, and the person answers (ADR 0072).
	Suggested bool `json:"suggested,omitempty"`
	// Many allows the choice next to other choices of its axis that are Many
	// too. By default a selection has one choice per axis.
	Many bool `json:"many,omitempty"`
	// Requires are choices that must be in the selection.
	Requires []string `json:"requires,omitempty"`
	// Conflicts are choices that must not be in the selection.
	Conflicts []string `json:"conflicts,omitempty"`
}

// Plays reports whether the choice's plugin plays the role in a stack built
// from it.
func (c Choice) Plays(role Role) bool { return slices.Contains(c.Roles, role) }
