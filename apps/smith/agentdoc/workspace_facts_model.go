package agentdoc

// Facts are the rules of a workspace that the agent file states. Collect fills
// them; Render only reads them, so the same Facts give the same text.
type Facts struct {
	// Name is the workspace name.
	Name string
	// Architectures are the architectures in force in at least one project, sorted.
	Architectures []Architecture
	// Languages are the languages of the projects, sorted, each once.
	Languages []string
	// Projects are the projects, sorted by root.
	Projects []Project
	// Tasks are the done-tasks (lint, typecheck, test, build) that at least one
	// project has, in that order.
	Tasks []string
	// Derive tells whether derived files are on.
	Derive bool
	// Derived are the paths smith owns, from smith.lock, sorted.
	Derived []string
	// Ejected are the derived paths a person owns, sorted.
	Ejected []string
	// Branch is the release branch.
	Branch string
}

// Architecture is an architecture name and what it means.
type Architecture struct {
	Name, Meaning string
	// Rules is the one-line summary of the checker's rules, and Layers its
	// parts from the inside out: both come from the ruleset `smith lint`
	// enforces, so the agent file and the check cannot disagree. Empty for an
	// architecture with no layers (vertical-slices).
	Rules  string
	Layers []Layer
}

// Layer is one part of an architecture as the checker knows it.
type Layer struct {
	Name, Meaning string
	Folders       []string
}

// Project is one project and the choices it was made with.
type Project struct {
	Name, Root, Kind, Language, Stack string
	// Choices are the choices of the project's stack, in the order of the stack.
	Choices []string
}
