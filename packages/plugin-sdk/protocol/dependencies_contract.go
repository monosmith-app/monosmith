package protocol

// InstallParams asks a plugin to add packages to one of its projects with
// the ecosystem's own toolchain, or, with no project, to install every
// dependency of the projects it serves (a fresh clone's first step).
type InstallParams struct {
	// Project is the project the packages go into; nil installs everything.
	Project *Project `json:"project,omitempty"`
	// Projects are the projects the plugin serves, when Project is nil.
	Projects []Project `json:"projects,omitempty"`
	// Packages are names or pinned specs as the ecosystem writes them
	// (left-pad, requests>=2.31, github.com/pkg/errors@v0.9.1).
	Packages []string `json:"packages,omitempty"`
	// Dev adds them as development dependencies where the ecosystem has
	// such a thing; a plugin without one says so in a note.
	Dev bool `json:"dev,omitempty"`
	// Quiet keeps the package manager's output unless it fails: the host sets
	// it when it installs as one step of a larger job (smith add).
	Quiet bool `json:"quiet,omitempty"`
}

// UninstallParams asks a plugin to remove packages from one of its projects.
type UninstallParams struct {
	Project  Project  `json:"project"`
	Packages []string `json:"packages"`
}

// DependencyResult lists the workspace-relative files the plugin changed and
// notes for the user.
type DependencyResult struct {
	Files []string `json:"files,omitempty"`
	Notes []string `json:"notes,omitempty"`
}

// LinkParams asks a plugin to make Project depend on Library, a lib or package
// of the same workspace and language, in the ecosystem's own form, and to
// install or lock as that needs. The host has checked that both exist, that
// the languages match, that Library is not an app, and that the link makes no
// cycle; the plugin answers with an error that names the file when it cannot
// write the dependency.
type LinkParams struct {
	// Project is the project that gains the dependency.
	Project Project `json:"project"`
	// Library is the project it depends on.
	Library Project `json:"library"`
	// Quiet keeps the package manager's output unless it fails.
	Quiet bool `json:"quiet,omitempty"`
	// Plan asks for the plan only: the plugin writes nothing and answers with
	// the manifest, the line it would write and the choices the user has.
	Plan bool `json:"plan,omitempty"`
	// Spec is the version range the user chose from LinkResult.Options; empty
	// means the plugin's default, or the form the workspace already uses.
	Spec string `json:"spec,omitempty"`
}

// LinkResult is what a link wrote, or with LinkParams.Plan would write.
type LinkResult struct {
	// Files are the workspace-relative files changed (or to be changed).
	Files []string `json:"files,omitempty"`
	// Notes are for the user.
	Notes []string `json:"notes,omitempty"`
	// Manifest is the file the dependency goes into, workspace-relative.
	Manifest string `json:"manifest,omitempty"`
	// Line is the dependency as it is written there, for the plan shown
	// before anything is written ("@acme/core": "*").
	Line string `json:"line,omitempty"`
	// Options are the real choices of the range written, with the default
	// first. Empty when the ecosystem has none, or the workspace already
	// follows a convention (then Line shows it).
	Options []LinkOption `json:"options,omitempty"`
}

// LinkOption is one way to write the dependency, with its consequence in
// words a person who read no manual can follow.
type LinkOption struct {
	// Spec is the value passed back in LinkParams.Spec.
	Spec string `json:"spec"`
	// Label is the option in a few words.
	Label string `json:"label"`
	// Meaning says what happens with this option.
	Meaning string `json:"meaning"`
}

// UnlinkParams asks a plugin to remove the dependency of Project on Library.
// Notes of the result say so when the manifest held no such dependency.
type UnlinkParams struct {
	Project Project `json:"project"`
	Library Project `json:"library"`
}

// DeclaredParams asks for the dependencies the plugin's projects declare
// (and the workspace root's manifest, when the ecosystem has one).
type DeclaredParams struct {
	Projects []Project `json:"projects"`
}

// Declaration is one dependency one manifest declares.
type Declaration struct {
	// Project is the declaring project's name, "" for the root manifest.
	Project string `json:"project,omitempty"`
	// File is the manifest, workspace-relative.
	File string `json:"file"`
	// Name is the dependency's name as the ecosystem compares names.
	Name string `json:"name"`
	// Spec is the declared range or version (^1.2.3, >=2.31), "" when the
	// manifest gives none.
	Spec string `json:"spec"`
	// Section is where it is declared: dependencies, devDependencies, dev...
	Section string `json:"section"`
	// Peer declarations state what a package works with, not what is
	// installed: sync and up leave them alone.
	Peer bool `json:"peer,omitempty"`
	// Rewritable is false for a spec smith must not rewrite: a URL, a git
	// ref, a workspace alias, a compound range, environment markers.
	Rewritable bool `json:"rewritable"`
}

// DeclaredResult is every declaration, and the version installed of each
// name where the plugin can tell.
type DeclaredResult struct {
	Declarations []Declaration     `json:"declarations"`
	Installed    map[string]string `json:"installed,omitempty"`
}

// RewriteParams asks a plugin to change declared specs in place.
type RewriteParams struct {
	Changes []Rewrite `json:"changes"`
}

// Rewrite changes one declaration's spec.
type Rewrite struct {
	Declaration Declaration `json:"declaration"`
	Spec        string      `json:"spec"`
}

// LatestParams asks for the newest stable release of each dependency, and,
// with Tools, how the tools the plugin pins compare with their latest.
type LatestParams struct {
	Names []string `json:"names,omitempty"`
	Tools bool     `json:"tools,omitempty"`
}

// LatestResult maps each name to its newest stable version. A name the
// registry could not be asked about is in Unreachable, with the reason; a
// name the registry does not have is in neither.
type LatestResult struct {
	Latest      map[string]string `json:"latest,omitempty"`
	Unreachable map[string]string `json:"unreachable,omitempty"`
	Tools       []ToolPin         `json:"tools,omitempty"`
}

// ToolPin is a tool the plugin installs at a pinned version (in CI, the dev
// container, the venv) against the newest release; informational.
type ToolPin struct {
	Name   string `json:"name"`
	Pinned string `json:"pinned"`
	// Latest is empty when its release feed could not be reached.
	Latest string `json:"latest,omitempty"`
	// Newer is true when Latest is a release the pin does not cover.
	Newer bool `json:"newer,omitempty"`
}
