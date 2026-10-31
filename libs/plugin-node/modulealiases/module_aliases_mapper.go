package modulealiases

import (
	"fmt"
	"maps"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Aliases are the import aliases of a project, each mapped to a directory
// relative to the project with no leading "./" and no trailing "/".
type Aliases map[string]string

// FromFacts reads module.aliases from the facts of a request. An absent fact,
// or one of the wrong shape, gives no aliases: the tools keep their defaults.
// An empty directory means the project root.
func FromFacts(facts map[string]any) Aliases {
	out := Aliases{}
	switch raw := facts[protocol.FactModuleAliases].(type) {
	case map[string]any:
		for alias, dir := range raw {
			if s, ok := dir.(string); ok && alias != "" {
				out[alias] = clean(s)
			}
		}
	case map[string]string:
		for alias, dir := range raw {
			if alias != "" {
				out[alias] = clean(dir)
			}
		}
	}
	return out
}

func clean(dir string) string {
	dir = strings.Trim(path.Clean(strings.ReplaceAll(dir, `\`, "/")), "/")
	if dir == "." {
		return ""
	}
	return dir
}

func (a Aliases) names() []string { return slices.Sorted(maps.Keys(a)) }

func target(dir, suffix string) string {
	if dir == "" {
		return "." + suffix
	}
	return "./" + dir + suffix
}

// TSConfigPaths is the compilerOptions "paths" member to append after the
// last member of a tsconfig, `,` and a new line included, or "" for no
// aliases. Each alias covers itself and everything under it.
func (a Aliases) TSConfigPaths() string {
	if len(a) == 0 {
		return ""
	}
	var lines []string
	for _, alias := range a.names() {
		dir := a[alias]
		lines = append(lines,
			fmt.Sprintf("      %q: [%q]", alias, target(dir, "")),
			fmt.Sprintf("      %q: [%q]", alias+"/*", target(dir, "/*")))
	}
	return ",\n    \"paths\": {\n" + strings.Join(lines, ",\n") + "\n    }"
}

// CompilerPaths is the compilerOptions "paths" value as a decoded JSON
// object, for a config a plugin builds in code (ts-jest's inline options), or
// nil for no aliases. It maps the same names as TSConfigPaths.
func (a Aliases) CompilerPaths() map[string]any {
	if len(a) == 0 {
		return nil
	}
	out := map[string]any{}
	for alias, dir := range a {
		out[alias] = []any{target(dir, "")}
		out[alias+"/*"] = []any{target(dir, "/*")}
	}
	return out
}

var regexpMeta = regexp.MustCompile(`[.*+?^${}()|\[\]\\]`)

// JestModuleNameMapper is the Jest moduleNameMapper entries of the aliases:
// each alias, and each path below it, to the directory under <rootDir>.
func (a Aliases) JestModuleNameMapper() map[string]string {
	out := map[string]string{}
	for alias, dir := range a {
		base := "<rootDir>"
		if dir != "" {
			base += "/" + dir
		}
		pattern := regexpMeta.ReplaceAllString(alias, `\$0`)
		out["^"+pattern+"$"] = base
		out["^"+pattern+"/(.*)$"] = base + "/$1"
	}
	return out
}

// ViteResolveAlias is the Vite resolve.alias member for a vite.config.ts or a
// vitest.config.ts, with its trailing new line, or "" for no aliases. A
// replacement that starts with "/" is relative to the project root, which is
// where Vite and Vitest run.
func (a Aliases) ViteResolveAlias() string {
	if len(a) == 0 {
		return ""
	}
	var lines []string
	for _, alias := range a.names() {
		lines = append(lines, fmt.Sprintf("      %q: %q,", alias, "/"+a[alias]))
	}
	return "  resolve: {\n    alias: {\n" + strings.Join(lines, "\n") + "\n    },\n  },\n"
}
