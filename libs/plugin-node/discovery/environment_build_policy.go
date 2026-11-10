package discovery

import (
	"regexp"
	"strings"
)

// environmentBuild matches the script of an environment build: vite build with
// a --mode and an --outDir (ADR 0081).
var environmentBuild = regexp.MustCompile(`\bvite\s+build\b`)

// outDirFlag reads the value of a --outDir flag.
var outDirFlag = regexp.MustCompile(`--outDir[= ]+"?([^"\s&;|]+)"?`)

// environmentBuildDir reports the folder a web app's environment build writes,
// when the script is one: the script is "build" or "build:<env>", runs
// vite build, and names its --outDir. The folder is relative to the project.
func environmentBuildDir(script, text string) (string, bool) {
	if script != "build" && !strings.HasPrefix(script, "build:") {
		return "", false
	}
	if !environmentBuild.MatchString(text) {
		return "", false
	}
	m := outDirFlag.FindStringSubmatch(text)
	if m == nil {
		return "", false
	}
	return m[1], true
}

// environmentOf is the environment a build script is for: "uat" for
// "build:uat"; "default" for "build".
func environmentOf(script string) string {
	if env, ok := strings.CutPrefix(script, "build:"); ok {
		return env
	}
	return "default"
}
