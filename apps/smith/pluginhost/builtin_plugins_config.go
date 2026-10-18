package pluginhost

import (
	"sort"

	bicepp "github.com/monosmith-app/monosmith/libs/plugin-bicep/plugin"
	cppp "github.com/monosmith-app/monosmith/libs/plugin-cpp/plugin"
	dartp "github.com/monosmith-app/monosmith/libs/plugin-dart/plugin"
	dockerp "github.com/monosmith-app/monosmith/libs/plugin-docker/plugin"
	dotnetp "github.com/monosmith-app/monosmith/libs/plugin-dotnet/plugin"
	gop "github.com/monosmith-app/monosmith/libs/plugin-go/plugin"
	"github.com/monosmith-app/monosmith/libs/plugin-go/pluginkit"
	jvmp "github.com/monosmith-app/monosmith/libs/plugin-jvm/plugin"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/angular"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/astro"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/eslint"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/jest"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/playwright"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/typescript"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vite"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vitest"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vscode"
	nodep "github.com/monosmith-app/monosmith/libs/plugin-node/plugin"
	phpp "github.com/monosmith-app/monosmith/libs/plugin-php/plugin"
	pythonp "github.com/monosmith-app/monosmith/libs/plugin-python/plugin"
	recipep "github.com/monosmith-app/monosmith/libs/plugin-recipe/plugin"
	rubyp "github.com/monosmith-app/monosmith/libs/plugin-ruby/plugin"
	rustp "github.com/monosmith-app/monosmith/libs/plugin-rust/plugin"
	slicesp "github.com/monosmith-app/monosmith/libs/plugin-vertical-slices/plugin"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// Builtins constructs each built-in language plugin by name.
var Builtins = map[string]func(version string) serve.Plugin{
	"node":   func(v string) serve.Plugin { return nodep.New(v) },
	"go":     func(v string) serve.Plugin { return gop.New(v) },
	"dotnet": func(v string) serve.Plugin { return dotnetp.New(v) },
	"ruby":   func(v string) serve.Plugin { return rubyp.New(v) },
	"php":    func(v string) serve.Plugin { return phpp.New(v) },
	"cpp":    func(v string) serve.Plugin { return cppp.New(v) },
	"dart":   func(v string) serve.Plugin { return dartp.New(v) },
	"jvm":    func(v string) serve.Plugin { return jvmp.New(v) },
	"bicep":  func(v string) serve.Plugin { return bicepp.New(v) },
	"python": func(v string) serve.Plugin { return pythonp.New(v) },
	"rust":   func(v string) serve.Plugin { return rustp.New(v) },
	"recipe": func(v string) serve.Plugin { return recipep.New(v) },
}

// BuiltinAddons constructs each built-in plugin that is not a language:
// toolchain, structure, scaffold and deploy plugins. They are always loaded,
// so their stacks are available, but act only on projects whose stack names
// them, or as a convention smith.toml configures ([conventions.<name>]).
var BuiltinAddons = map[string]func(version string) serve.Plugin{
	slicesp.Name:    func(v string) serve.Plugin { return slicesp.New(v) },
	eslint.Name:     func(v string) serve.Plugin { return eslint.New(v) },
	jest.Name:       func(v string) serve.Plugin { return jest.New(v) },
	vitest.Name:     func(v string) serve.Plugin { return vitest.New(v) },
	playwright.Name: func(v string) serve.Plugin { return playwright.New(v) },
	typescript.Name: func(v string) serve.Plugin { return typescript.New(v) },
	vite.Name:       func(v string) serve.Plugin { return vite.New(v) },
	angular.Name:    func(v string) serve.Plugin { return angular.New(v) },
	astro.Name:      func(v string) serve.Plugin { return astro.New(v) },
	vscode.Name:     func(v string) serve.Plugin { return vscode.New(v) },
	pluginkit.Name:  func(v string) serve.Plugin { return pluginkit.New(v) },
	dockerp.Name:    func(v string) serve.Plugin { return dockerp.New(v) },
}

// Builtin returns the constructor of a built-in plugin.
func Builtin(name string) (func(version string) serve.Plugin, bool) {
	if ctor, ok := Builtins[name]; ok {
		return ctor, true
	}
	ctor, ok := BuiltinAddons[name]
	return ctor, ok
}

// BuiltinNames lists the built-in language plugins, sorted.
func BuiltinNames() []string { return sortedNames(Builtins) }

func sortedNames[V any](m map[string]V) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
