package webenv

import (
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// OutDir is the folder one environment builds into, relative to the app's
// folder: ../../dist/apps/<app>/<env> in the workspace, or dist/<env> inside
// the app. dir is the app's workspace-relative folder (apps/web).
func OutDir(s Settings, dir, env string) string {
	if s.Normalized().Output == OutputProject {
		return "dist/" + env
	}
	return protocol.BuildDirFromProject(dir) + "/" + env
}

// TaskName is the task that builds an environment: "build" for the default
// environment, "build:<env>" for each other one.
func TaskName(s Settings, env string) string {
	if s.Normalized().Default == env {
		return "build"
	}
	return "build:" + env
}

// Scripts are the package.json scripts that build each environment: one
// vite build per environment, with its own --mode and --outDir. The --outDir
// is what smith reads as the task's outputs, so the cache and the build agree
// on the folder. The start script runs the dev server in the lcl environment
// when the app has one.
//
// Parameters: s are the settings; dir is the app's workspace-relative folder.
func Scripts(s Settings, dir string) map[string]string {
	n := s.Normalized()
	out := map[string]string{}
	for _, env := range n.Names {
		out[TaskName(n, env)] = fmt.Sprintf("vite build --mode %s --outDir %s", env, OutDir(n, dir, env))
	}
	if slices.Contains(n.Names, "lcl") {
		out["start"] = "vite --mode lcl"
	}
	return out
}

// EnvFileName is the file Vite reads for an environment: .env.<env>.
func EnvFileName(env string) string { return ".env." + env }

// envHeader opens every env file: what the file is, and the one rule.
const envHeader = `# Values with the VITE_ prefix end up in the browser bundle: never put a secret here.
# A secret belongs to the server or to [ci] secrets in smith.toml, not to a web app.
`

// EnvFile is the starter content of the file of an environment.
func EnvFile(env string, s Settings) string {
	build := "smith run " + TaskName(s, env)
	return fmt.Sprintf("# Environment %s: read by `%s` (vite build --mode %s).\n%sVITE_API_URL=\n", env, build, env, envHeader)
}

// Files are the files an app gets for its environments, by path relative to
// the app's folder: .env.<env> for each, .env.example when only the example
// is committed, the Vite build options (vite.build.ts) and a .gitignore that
// keeps .env.*.local (and, with FilesExample, the .env.<env> files) out of git.
func Files(s Settings) map[string]string {
	n := s.Normalized()
	files := map[string]string{"vite.build.ts": BuildOptionsFile(n), ".gitignore": GitIgnore(n)}
	for _, env := range n.Names {
		files[EnvFileName(env)] = EnvFile(env, n)
	}
	if n.EnvFiles == FilesExample {
		files[".env.example"] = fmt.Sprintf("# Copy to .env.<environment> (%s) and fill in the values.\n%sVITE_API_URL=\n", strings.Join(n.Names, ", "), envHeader)
	}
	return files
}

// GitIgnore is the app's .gitignore for the environment files.
func GitIgnore(s Settings) string {
	lines := []string{"# Private overrides of an environment file: never committed.", ".env.*.local"}
	if s.Normalized().EnvFiles == FilesExample {
		lines = append(lines, "# Only .env.example is committed: each machine and CI job makes its own .env.<environment>.", ".env.*", "!.env.example")
	}
	return strings.Join(lines, "\n") + "\n"
}

// BuildOptionsFile is vite.build.ts: the function build(mode) that returns the
// Vite build options of an environment. lcl keeps sourcemaps and does not
// minify; dev does not minify; the others minify with the minifier that
// Vite ships with (terser would need the extra dependency terser).
func BuildOptionsFile(s Settings) string {
	var b strings.Builder
	b.WriteString(`import type { BuildOptions } from 'vite'

/**
 * The Vite build options of one environment. vite.config.ts calls it with the mode of
 * ` + "`vite build --mode <environment>`" + `. smith builds each environment with its own task:
 * the default one with ` + "`smith run build`" + `, the others with ` + "`smith run build:<environment>`" + `.
 * The output folder is the --outDir of that script in package.json, so the build and
 * smith's cache name the same folder.
 */
export function build (mode: string): BuildOptions {
  const options: BuildOptions = {
    emptyOutDir: true,
    manifest:    false,
    sourcemap:   false,
    // true is the minifier Vite ships with; 'terser' needs the extra dependency terser.
    minify:      true,
  }

  switch (mode) {
`)
	n := s.Normalized()
	if slices.Contains(n.Names, "lcl") {
		b.WriteString("    case 'lcl': {\n      options.sourcemap = true\n      options.minify = false\n      break\n    }\n")
	}
	if slices.Contains(n.Names, "dev") {
		b.WriteString("    case 'dev': {\n      options.minify = false\n      break\n    }\n")
	}
	b.WriteString("    default: {\n      break\n    }\n  }\n\n  return options\n}\n")
	return b.String()
}

// ViteConfigTemplate is vite.config.ts for an app with environments: the
// plain template of the plugin (see base.toml, with the same slots) with the
// build options taken from build(mode). It uses the plugin's [[ ]] delimiters.
const ViteConfigTemplate = `[[range index .Slots "vite.imports"]][[.]]
[[end]]import { defineConfig } from 'vite'
import { build } from './vite.build.ts'

export default defineConfig(({ mode }) => ({
[[if index .Slots "vite.plugins"]]  plugins: [[ "[" ]][[range $i, $p := index .Slots "vite.plugins"]][[if $i]], [[end]][[$p]][[end]]],
[[end]][[.ViteAlias]]  build:[[if and (index .Slots "vite.plugins") (not .ViteAlias)]]   [[else]] [[end]]build(mode),
}))
`
