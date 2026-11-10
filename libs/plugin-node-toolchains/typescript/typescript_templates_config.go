package typescript

import (
	"maps"

	"github.com/monosmith-app/monosmith/libs/plugin-node/scaffolding"
)

// templates is the TypeScript starter the plugin scaffolds from.
var templates = scaffolding.TemplateSet{
	Slice: sliceTemplates, App: appTemplates, Lib: libTemplates, Package: packageTemplates,
	Root: rootTemplates, DevDependencies: devDependencies,
}

// devDependencies are the root TypeScript tooling a workspace pins. Nothing
// installs "latest"; bump these deliberately. TypeScript stays on 6.x until
// ts-jest and typescript-eslint support 7.
var devDependencies = map[string]string{
	"@types/node":         "^24.19.2",
	"typescript":          "~6.0.3",
	"@monosmith/tsconfig": "^0.1.0",
}

// esbuildVersion is the esbuild a project that follows siblings building with
// it pins in the workspace.
const esbuildVersion = "^0.25.12"

// GreetingSlice is the greeting slice every TypeScript project starts with:
// a contract, a use case and the slice's index (its only public door). A
// test-runner plugin (jest, vitest) adds the use case's spec.
var GreetingSlice = map[string]string{
	"[[.Dir]]/src/greeting/greeting.contract.ts": `/** A greeting addressed to someone. */
export interface Greeting {
  message: string
}
`,
	"[[.Dir]]/src/greeting/greet.use-case.ts": `import type { Greeting } from './greeting.contract.ts'

/**
 * Greets someone by name.
 * @param name - who to greet
 * @returns the greeting
 */
export function greet (name: string): Greeting {
  return { message: ` + "`Hello, ${name}!`" + ` }
}
`,
	"[[.Dir]]/src/greeting/index.ts": `export { greet } from './greet.use-case.ts'
export type { Greeting } from './greeting.contract.ts'
`,
}

// sliceTemplates are the greeting slice and the project's tsconfigs. The
// project's tsconfig sets rewriteRelativeImportExtensions itself: the starter
// imports './greeting/index.ts' (Node's ESM needs a file, not a directory),
// and a base smith did not write, such as a converted Nx workspace's, may not
// allow .ts specifiers.
var sliceTemplates = withGreetingSlice(map[string]string{
	"[[.Dir]]/tsconfig.json": `[[if eq .BuildConfig "tsconfig.lib.json"]]{
  "extends": "../../tsconfig.base.json",
  "files": [],
  "include": [],
  "references": [
    { "path": "./tsconfig.lib.json" }[[if .TsconfigSpec]],
    { "path": "./tsconfig.spec.json" }[[end]]
  ]
}
[[else]]` + projectTsconfig + `[[end]]`,
	"[[.Dir]]/tsconfig.build.json": `{
  "extends": "./tsconfig.json",
  "exclude": ["src/**/*.[[.TestSuffix]].ts"]
}
`,
})

// projectTsconfig is the compiler configuration of a project: the content of
// tsconfig.json, or of tsconfig.lib.json when tsconfig.json only references it.
const projectTsconfig = `{
  "extends": "../../tsconfig.base.json",
  "compilerOptions": {
    "rootDir": "src",
    "rewriteRelativeImportExtensions": true,
    "types": ["node"][[.TSPaths]]
  },
  "include": ["src"]
}
`

// libTsconfig is the tsconfig.lib.json of a project whose siblings have one: the
// build configuration, which tsconfig.json references.
const libTsconfig = `{
  "extends": "../../tsconfig.base.json",
  "compilerOptions": {
    "rootDir": "src",
    "rewriteRelativeImportExtensions": true,
    "types": ["node"][[.TSPaths]]
  },
  "include": ["src"],
  "exclude": ["src/**/*.[[.TestSuffix]].ts"]
}
`

// specTsconfig is the tsconfig.spec.json of a project whose siblings have one.
const specTsconfig = `{
  "extends": "../../tsconfig.base.json",
  "compilerOptions": {
    "rootDir": "src",
    "rewriteRelativeImportExtensions": true,
    "noEmit": true,
    "types": ["node"][[.TSPaths]]
  },
  "include": ["src/**/*.[[.TestSuffix]].ts"]
}
`

// setFor is the template set for a shape: the starter, with the tsconfig files
// the siblings have. A project whose siblings have a tsconfig.lib.json gets it in
// place of tsconfig.build.json, and a tsconfig.spec.json when they have one.
func setFor(shape map[string]string) scaffolding.TemplateSet {
	set := templates
	build, spec := scaffolding.TsconfigSet(shape)
	if build == scaffolding.TsconfigLib {
		set.Slice = maps.Clone(templates.Slice)
		delete(set.Slice, "[[.Dir]]/"+scaffolding.TsconfigBuild)
		set.Slice["[[.Dir]]/"+scaffolding.TsconfigLib] = libTsconfig
	}
	if spec {
		set.Slice = maps.Clone(set.Slice)
		set.Slice["[[.Dir]]/"+scaffolding.TsconfigSpec] = specTsconfig
	}
	return set
}

// withGreetingSlice adds the greeting slice to a project's templates.
func withGreetingSlice(templates map[string]string) map[string]string {
	maps.Copy(templates, GreetingSlice)
	return templates
}

// shapeVars are the template variables the package.json of a project reads:
// esbuild builds with esbuild, dist leads the entry points to the build folder,
// and build says the project has a build script (a lib has none unless the
// shape it follows asks for a build).
const shapeVars = `[[- $esbuild := eq (index .Shape "build.tool") "esbuild" -]]
[[- $dist := eq (index .Shape "exports.target") "dist" -]]
[[- $build := or $esbuild $dist (eq (index .Shape "build.tool") "tsc") -]]
[[- $map := eq (index .Shape "exports.map") "map" -]]
`

// buildRef is the build folder as package.json names it.
const buildRef = `[[if .BuildLocal]]./[[end]][[.Build]]`

// entryPoints are the fields that lead to the code: the build when the siblings
// export their build (with an exports map when they have one), else the source.
const entryPoints = `[[if $dist]]  "main": "` + buildRef + `/index.js",
  "types": "` + buildRef + `/index.d.ts",
[[if $map]]  "exports": {
    "./package.json": "./package.json",
    ".": {
      "types": "` + buildRef + `/index.d.ts",
      "import": "` + buildRef + `/index.js",
      "default": "` + buildRef + `/index.js"
    }
  },
[[end]][[if .BuildLocal]]  "files": ["dist"],
[[end]][[else]]  "exports": "./src/index.ts",
[[end]]`

// typecheckProject is the tsconfig file the typecheck script reads: the one that
// has the compiler options.
const typecheckProject = `[[if eq .BuildConfig "tsconfig.lib.json"]]tsconfig.lib.json[[else]]tsconfig.json[[end]]`

// libraryBuild is the build script of a lib or package: tsc, or esbuild for the
// code and tsc for the declarations.
const libraryBuild = `[[if $esbuild]]esbuild src/index.ts --bundle --packages=external --platform=node --format=esm --outdir=[[.Build]] && tsc -p [[.BuildConfig]] --declaration --emitDeclarationOnly [[if .BuildLocal]]--declarationDir[[else]]--outDir[[end]] [[.Build]][[else]]tsc -p [[.BuildConfig]][[if $dist]] --declaration[[end]] --outDir [[.Build]][[end]]`

var appTemplates = map[string]string{
	"[[.Dir]]/package.json": shapeVars + `{
  "name": "[[.Package]]",
  "version": "0.0.0",
  "private": true,
  "type": "module",
  "smith": { "kind": "app"[[if .Stack]], "stack": "[[.Stack]]"[[end]] },
  "main": "[[.Build]]/main.js",
  "scripts": {
    "build": "[[if $esbuild]]esbuild src/main.ts --bundle --packages=external --platform=node --format=esm --outdir=[[.Build]][[else]]tsc -p [[.BuildConfig]] --outDir [[.Build]][[end]]",
    "typecheck": "tsc -p ` + typecheckProject + ` --noEmit",
    "start": "node [[.Build]]/main.js"
  }
}
`,
	"[[.Dir]]/src/main.ts": `import { greet } from './greeting/index.ts'

process.stdout.write(greet('world').message + '\n')
`,
}

var libTemplates = map[string]string{
	"[[.Dir]]/package.json": shapeVars + `{
  "name": "[[.Package]]",
  "version": "0.0.0",
  "private": true,
  "type": "module",
  "smith": { "kind": "lib"[[if .Stack]], "stack": "[[.Stack]]"[[end]] },
` + entryPoints + `  "scripts": {
[[if $build]]    "build": "` + libraryBuild + `",
[[end]]    "typecheck": "tsc -p ` + typecheckProject + ` --noEmit"
  }
}
`,
	"[[.Dir]]/src/index.ts": `export { greet } from './greeting/index.ts'
export type { Greeting } from './greeting/index.ts'
`,
}

var packageTemplates = map[string]string{
	"[[.Dir]]/package.json": shapeVars + `{
  "name": "[[.Package]]",
  "version": "0.0.0",
  "description": "[[.Name]]",
  "license": "MIT",
  "type": "module",
  "smith": { "kind": "package"[[if .Stack]], "stack": "[[.Stack]]"[[end]] },
` + entryPoints + `  "publishConfig": { "access": "public" },
  "scripts": {
    "build": "` + libraryBuild + `",
    "typecheck": "tsc -p ` + typecheckProject + ` --noEmit"
  }
}
`,
	"[[.Dir]]/src/index.ts": `export { greet } from './greeting/index.ts'
export type { Greeting } from './greeting/index.ts'
`,
	"[[.Dir]]/README.md": `# [[.Package]]

Generated by smith. Released automatically from conventional commits.
`,
}

// Root files written once, by the first TypeScript project.
var rootTemplates = map[string]string{
	"tsconfig.base.json": `{
  "extends": "@monosmith/tsconfig/node.json"
}
`,
}
