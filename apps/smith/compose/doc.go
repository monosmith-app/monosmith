// Package compose is `smith compose`: one catalogue of everything smith can
// put in a project (languages, frameworks, desktop shells, cloud function
// hosts, test, lint and style tools, architecture and CI plugins), a search
// over it, a check of the picks that names every conflict and every open
// question, and the plan that the picks end in: the exact `smith add` line,
// the files it writes and the tools it needs (ADR 0076).
//
// The catalogue is built from what the loaded plugins publish (their choices
// and roles), not from a list kept here. The screens (terminal, VS Code) show
// this package's JSON and add no rules of their own.
package compose
