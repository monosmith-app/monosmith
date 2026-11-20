package vite

import (
	"embed"

	"github.com/monosmith-app/monosmith/libs/plugin-node/choicedata"
)

// choiceFiles are the choices of the plugin as data (ADR 0045): the frameworks
// and the styles of a browser app, the version catalog they name packages
// from, the base they build on and the stacks the plugin publishes. To add a
// framework or a style, add a file here.
//
//go:embed choices
var choiceFiles embed.FS

// webChoices is the checked set of the embedded files.
var webChoices = choicedata.MustLoad(choiceFiles, "choices")

// ReactBuildPackages are the development packages that build a React app with
// Vite (vite and @vitejs/plugin-react), with the versions of the catalog. A
// tool that moves an existing React app onto Vite installs these, so it pins
// the same versions as a new app, and the npm overrides a new app's workspace
// gets (the Babel that @vitejs/plugin-react may peer, issue 584).
//
// Returns: new maps of package name to version range (the development
// packages, then the overrides); the caller may change them.
func ReactBuildPackages() (devDependencies, overrides map[string]string) {
	devDependencies, overrides = map[string]string{}, map[string]string{}
	plan, err := webChoices.Plan([]string{"react"})
	if err != nil {
		return devDependencies, overrides
	}
	for _, name := range []string{"vite", "@vitejs/plugin-react"} {
		if v, ok := plan.DevDependencies[name]; ok {
			devDependencies[name] = v
		}
	}
	for name, v := range plan.Overrides {
		overrides[name] = v
	}
	return devDependencies, overrides
}
