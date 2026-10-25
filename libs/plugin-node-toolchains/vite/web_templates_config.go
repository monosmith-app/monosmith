package vite

import (
	"maps"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/typescript"
)

// devDependencies every web app's tooling needs, at the workspace root.
// Nothing installs "latest"; bump these deliberately.
var devDependencies = map[string]string{
	"typescript": "~6.0.3",
	"vite":       "^8.3.4",
}

// webApp is one stack's starter: the files beyond the shared ones, and what
// it adds to the workspace root and the app's dependencies.
type webApp struct {
	// Entry is the module index.html loads, under src/.
	Entry string
	// Typecheck is the app's typecheck script.
	Typecheck string
	// Plugin is the Vite plugin's import and call, empty for none.
	PluginImport, PluginCall string
	// JSX compiles .tsx with React's runtime.
	JSX bool
	// Components are the extensions of single-file components to check.
	Components      []string
	Dependencies    map[string]string
	DevDependencies map[string]string
	// Overrides are npm overrides the app needs in the workspace root.
	Overrides map[string]string
	Files     map[string]string
}

// templates are the app's files: the greeting slice, the manifest,
// index.html, the Vite and TypeScript configs, and the stack's own files.
func (a webApp) templates() map[string]string {
	files := maps.Clone(typescript.GreetingSlice)
	maps.Copy(files, a.Files)
	files["[[.Dir]]/package.json"] = a.manifest()
	files["[[.Dir]]/index.html"] = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>[[.Name]]</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/` + a.Entry + `"></script>
  </body>
</html>
`
	files["[[.Dir]]/vite.config.ts"] = a.viteConfig()
	files["[[.Dir]]/tsconfig.json"] = a.tsconfig()
	return files
}

func (a webApp) manifest() string {
	var deps []string
	for _, name := range slices.Sorted(maps.Keys(a.Dependencies)) {
		deps = append(deps, `    "`+name+`": "`+a.Dependencies[name]+`"`)
	}
	dependencies := ""
	if len(deps) > 0 {
		dependencies = `,
  "dependencies": {
` + strings.Join(deps, ",\n") + `
  }`
	}
	return `{
  "name": "[[.Package]]",
  "version": "0.0.0",
  "private": true,
  "type": "module",
  "smith": { "kind": "app"[[if .Stack]], "stack": "[[.Stack]]"[[end]] },
  "scripts": {
    "build": "vite build",
    "typecheck": "` + a.Typecheck + `",
    "start": "vite"
  }` + dependencies + `
}
`
}

// viteConfig builds into the workspace's dist/<app dir>, in the layout
// @monosmith/eslint-config formats (aligned values).
func (a webApp) viteConfig() string {
	imports, plugins, build := "", "", "build: {"
	if a.PluginCall != "" {
		imports = a.PluginImport + "\n"
		plugins = "  plugins: [" + a.PluginCall + "],\n"
		build = "build:   {"
	}
	return imports + `import { defineConfig } from 'vite'

// The build goes to the workspace's dist/ (smith), never dist/ in the app.
export default defineConfig({
` + plugins + `  ` + build + `
    outDir:      '[[.Build]]',
    emptyOutDir: true,
  },
})
`
}

// tsconfig type-checks for the browser through Vite: bundler resolution,
// nothing emitted (Vite builds).
func (a webApp) tsconfig() string {
	jsx := ""
	if a.JSX {
		jsx = `
    "jsx": "react-jsx",`
	}
	include := []string{`"src/**/*.ts"`}
	if a.JSX {
		include = append(include, `"src/**/*.tsx"`)
	}
	for _, ext := range a.Components {
		include = append(include, `"src/**/*.`+ext+`"`)
	}
	include = append(include, `"vite.config.ts"`)
	return `{
  "compilerOptions": {
    "target": "es2023",
    "lib": ["es2023", "dom", "dom.iterable"],
    "module": "esnext",
    "moduleResolution": "bundler",` + jsx + `
    "strict": true,
    "isolatedModules": true,
    "skipLibCheck": true,
    "noEmit": true,
    "allowImportingTsExtensions": true,
    "verbatimModuleSyntax": true,
    "forceConsistentCasingInFileNames": true,
    "types": ["vite/client"]
  },
  "include": [` + strings.Join(include, ", ") + `]
}
`
}
