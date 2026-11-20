package projecttooling

import (
	"path"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/eslintpeer"
)

// Facts are what a project's files say about its build and lint tools.
type Facts struct {
	// LegacyESLintFiles are the .eslintrc files of the project and of the
	// workspace root, workspace-relative.
	LegacyESLintFiles []string
	// ESLintRange is the eslint range the project (or the root) depends on,
	// "" when none.
	ESLintRange string
	// ESLintInManifest is true when package.json has an eslintConfig key.
	ESLintInManifest bool
	// BuildScript is the project's "build" script, "" when it has none.
	BuildScript string
	// OutDir is where a tsc build writes, project-relative; "" when unknown.
	OutDir string
	// Dependencies are the names the project depends on, in any section.
	Dependencies map[string]bool
}

// BuildKind is what a build script runs, as far as smith's build is concerned.
type BuildKind string

// The kinds of build.
const (
	// BuildNone is a project without a build script.
	BuildNone BuildKind = "none"
	// BuildSmith is a build smith writes itself: vite, esbuild, or tsc into dist.
	BuildSmith BuildKind = "smith"
	// BuildTsc is a tsc build into another folder (lib/, build/).
	BuildTsc BuildKind = "tsc"
	// BuildCRA is create-react-app (react-scripts build).
	BuildCRA BuildKind = "create-react-app"
	// BuildOther is a build smith does not rewrite (rollup, webpack, a script).
	BuildOther BuildKind = "other"
)

// bundlers are the build tools whose configuration smith does not rewrite.
var bundlers = []string{"rollup", "webpack", "parcel", "microbundle", "tsdx", "tsup", "babel", "gulp", "grunt"}

// ClassifyBuild says what kind of build the facts describe.
//
// Parameters: f are the project's facts.
// Returns: the kind.
func ClassifyBuild(f Facts) BuildKind {
	script := strings.TrimSpace(f.BuildScript)
	words := strings.Fields(script)
	switch {
	case script == "":
		return BuildNone
	case strings.Contains(script, "react-scripts build"):
		return BuildCRA
	case words[0] == "vite" || words[0] == "esbuild":
		return BuildSmith
	case words[0] == "tsc" && !strings.ContainsAny(script, "&|;"):
		if out := strings.TrimPrefix(path.Clean(f.OutDir), "./"); f.OutDir == "" || out == "dist" || strings.Contains(out, "dist/") {
			return BuildSmith
		}
		return BuildTsc
	}
	return BuildOther
}

// bundlerOf is the bundler a build script runs, "" when none of the known ones.
func bundlerOf(script string) string {
	for _, w := range strings.Fields(script) {
		for _, b := range bundlers {
			if w == b || strings.HasPrefix(w, b+"-") {
				return b
			}
		}
	}
	return ""
}

// LegacyReasons lists, in words, the tools of a project that smith's tooling
// replaces and that are legacy: an ESLint 8 configuration, create-react-app,
// a tsc build outside dist/, a bundler build. Empty means the project's own
// tools are current, and smith recommends keeping them.
//
// Parameters: f are the project's facts.
// Returns: the reasons, in a fixed order.
func LegacyReasons(f Facts) []string {
	var out []string
	if f.Dependencies["react-scripts"] || strings.Contains(f.BuildScript, "react-scripts") {
		out = append(out, "create-react-app (react-scripts), which is no longer maintained")
	}
	switch {
	case len(f.LegacyESLintFiles) > 0:
		out = append(out, "an ESLint 8 configuration ("+strings.Join(f.LegacyESLintFiles, ", ")+")")
	case f.ESLintInManifest:
		out = append(out, "an ESLint 8 configuration (eslintConfig in package.json)")
	case f.ESLintRange != "" && eslintMajor(f.ESLintRange) > 0 && eslintMajor(f.ESLintRange) < 9:
		out = append(out, "ESLint "+strconv.Itoa(eslintMajor(f.ESLintRange)))
	}
	switch ClassifyBuild(f) {
	case BuildTsc:
		out = append(out, "a tsc build into "+strings.TrimSuffix(strings.TrimPrefix(f.OutDir, "./"), "/")+"/")
	case BuildOther:
		if b := bundlerOf(f.BuildScript); b != "" {
			out = append(out, "a "+b+" build")
		}
	}
	return out
}

// Recommend is the answer smith recommends: its own tooling when a tool of
// the project is legacy, the project's tools otherwise.
//
// Parameters: reasons are the LegacyReasons of the project.
// Returns: Smith or Keep.
func Recommend(reasons []string) Answer {
	if len(reasons) > 0 {
		return Smith
	}
	return Keep
}

// eslintMajor is the major version of the lowest version an eslint range
// accepts, 0 when the range has no plain lowest version.
func eslintMajor(r string) int {
	floor, ok := eslintpeer.RangeFloor(r)
	if !ok {
		return 0
	}
	major, _, _ := strings.Cut(floor, ".")
	n, err := strconv.Atoi(major)
	if err != nil {
		return 0
	}
	return n
}
