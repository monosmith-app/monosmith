package plugin

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// nodeLanguages are the names a Node project's language goes by. Their slice
// rules are not a slice-check task: @monosmith/eslint-config enforces them in
// lint, so the convention only checks that the root ESLint config loads it.
var nodeLanguages = []string{"node", "javascript", "js"}

const eslintConfigPackage = "@monosmith/eslint-config"

var (
	eslintConfigNames = []string{"eslint.config.mjs", "eslint.config.js", "eslint.config.cjs", "eslint.config.ts", "eslint.config.mts", "eslint.config.cts"}
	slicesOptedOut    = regexp.MustCompile(`verticalSlices\s*:\s*false`)
	configImport      = regexp.MustCompile(`from\s+['"](@[\w.-]+/eslint-config|eslint-config-[\w.-]+)['"]|require\(\s*['"](@[\w.-]+/eslint-config|eslint-config-[\w.-]+)['"]\s*\)`)
)

// checkNodeESLint reports whether the workspace's root ESLint config loads
// @monosmith/eslint-config, which is where a Node project's vertical-slice
// rules come from. It returns nothing when no Node project is covered, or
// when the config opts out with verticalSlices: false.
func checkNodeESLint(workspaceRoot string, projects []protocol.Project) []protocol.Finding {
	var covered []string
	for _, p := range projects {
		if slices.Contains(nodeLanguages, p.Language) {
			covered = append(covered, p.Name)
		}
	}
	if len(covered) == 0 {
		return nil
	}
	finding := protocol.Finding{
		Check:  "Node projects are checked for vertical slices (" + eslintConfigPackage + ")",
		Remedy: "make the root eslint.config.mjs:\n  import monosmith from '" + eslintConfigPackage + "'\n  export default monosmith({ workspaceRoot: import.meta.dirname })\nand add " + eslintConfigPackage + " to the root devDependencies (npm i -D " + eslintConfigPackage + "). It also turns on Standard-style formatting and type-aware rules, so try it on a copy first; to keep another config, drop the vertical-slices convention for these projects ([conventions.vertical-slices] exclude = [...])",
	}
	for _, name := range eslintConfigNames {
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, name))
		if err != nil {
			continue
		}
		text := string(raw)
		switch {
		case strings.Contains(text, eslintConfigPackage) && slicesOptedOut.MatchString(text):
			return nil
		case strings.Contains(text, eslintConfigPackage):
			finding.OK = true
		default:
			finding.Detail = name + " does not load " + eslintConfigPackage + usingInstead(text) + ", so the slice rules do not run on " + named(covered)
		}
		return []protocol.Finding{finding}
	}
	finding.Detail = "the workspace root has no eslint.config file, so the slice rules do not run on " + named(covered)
	return []protocol.Finding{finding}
}

// usingInstead names the shared ESLint config the file loads, for the message.
func usingInstead(text string) string {
	m := configImport.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	return " (it uses " + m[1] + m[2] + ")"
}

// named lists a few project names and counts the rest, so a large workspace
// does not print every project.
func named(names []string) string {
	const shown = 4
	if len(names) <= shown {
		return strings.Join(names, ", ")
	}
	return strings.Join(names[:shown], ", ") + " and " + strconv.Itoa(len(names)-shown) + " more"
}
