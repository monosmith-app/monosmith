package vite

import (
	"fmt"
	"maps"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vite/webenv"
	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
)

// withEnvironments turns the starter of a web app into one that is built once
// per environment (ADR 0081): the build script becomes one script per
// environment, vite.config.ts takes its build options from build(mode) in
// vite.build.ts, and each environment gets its .env.<env> file.
//
// Parameters: set is the starter; s the environments; dir the workspace
// relative folder of the app.
// Returns: the starter with the environment files; an error when the settings
// are not valid (a name that is not a lowercase word, a default that is not in
// the list).
func withEnvironments(set scaffolding.TemplateSet, s webenv.Settings, dir string) (scaffolding.TemplateSet, error) {
	if err := s.Validate(); err != nil {
		return set, fmt.Errorf("the environments of %s: %w", dir, err)
	}
	out := set
	out.Slice = maps.Clone(set.Slice)
	out.Slice["[[.Dir]]/vite.config.ts"] = webenv.ViteConfigTemplate
	for path, body := range webenv.Files(s) {
		out.Slice["[[.Dir]]/"+path] = body
	}
	out.Scripts = maps.Clone(set.Scripts)
	delete(out.Scripts, "build")
	maps.Copy(out.Scripts, webenv.Scripts(s, dir))
	return out, nil
}
