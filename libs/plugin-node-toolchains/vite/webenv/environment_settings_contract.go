// Package webenv is what a web app needs to be built once per environment
// (ADR 0081): the settings that describe the environments, the files and the
// scripts that follow from them, and the names it checks. The vite plugin
// writes them when it scaffolds an app, and smith environments changes them
// in an app that exists.
package webenv

import (
	"slices"
)

// Where the build of an environment goes.
const (
	// OutputWorkspace is dist/apps/<app>/<env> in the workspace (the default).
	OutputWorkspace = "workspace"
	// OutputProject is dist/<env> in the app's own folder.
	OutputProject = "project"
)

// What is committed of the environment files.
const (
	// FilesCommit commits .env.<env> (the default: the files hold no secrets).
	FilesCommit = "commit"
	// FilesExample commits only .env.example; .env.<env> stays on the machine.
	FilesExample = "example"
)

// PluginName is the plugin whose settings table holds the environments
// ([projects.<name>.settings.vite]).
const PluginName = "vite"

// DefaultNames are the environments an app starts from: lcl is the
// developer's machine, then dev, uat and prd.
var DefaultNames = []string{"lcl", "dev", "uat", "prd"}

// Settings describes the environment builds of one web app. It is the
// settings.vite table of the project in smith.toml.
type Settings struct {
	// Names are the environments, in the order they are listed.
	Names []string
	// Default is the environment the task "build" builds; the others get the
	// task "build:<name>".
	Default string
	// Output is OutputWorkspace or OutputProject.
	Output string
	// EnvFiles is FilesCommit or FilesExample.
	EnvFiles string
}

// Normalized fills what a caller left out: the default environment is prd, or
// the last name when there is no prd; the output goes to the workspace; the
// environment files are committed.
func (s Settings) Normalized() Settings {
	out := s
	out.Names = slices.Clone(s.Names)
	if out.Default == "" && len(out.Names) > 0 {
		out.Default = out.Names[len(out.Names)-1]
		if slices.Contains(out.Names, "prd") {
			out.Default = "prd"
		}
	}
	if out.Output == "" {
		out.Output = OutputWorkspace
	}
	if out.EnvFiles == "" {
		out.EnvFiles = FilesCommit
	}
	return out
}

// ToMap is the settings as the table smith.toml and the plugin protocol carry.
func (s Settings) ToMap() map[string]any {
	n := s.Normalized()
	names := make([]any, len(n.Names))
	for i, name := range n.Names {
		names[i] = name
	}
	return map[string]any{"environments": names, "default_environment": n.Default, "output": n.Output, "env_files": n.EnvFiles}
}

// FromMap reads the settings table of a project. ok is false when the table
// names no environment, which means the app has a single build.
func FromMap(m map[string]any) (s Settings, ok bool) {
	switch v := m["environments"].(type) {
	case []string:
		s.Names = slices.Clone(v)
	case []any:
		for _, e := range v {
			if name, isString := e.(string); isString {
				s.Names = append(s.Names, name)
			}
		}
	}
	if len(s.Names) == 0 {
		return Settings{}, false
	}
	s.Default, _ = m["default_environment"].(string)
	s.Output, _ = m["output"].(string)
	s.EnvFiles, _ = m["env_files"].(string)
	return s.Normalized(), true
}
