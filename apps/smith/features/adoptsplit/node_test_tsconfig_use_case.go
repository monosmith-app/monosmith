package adoptsplit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// buildTSConfig is the build configuration smith's own scaffolds write: the
// sources only, while tsconfig.json covers the sources and the tests.
const buildTSConfig = "tsconfig.build.json"

// plainTsc matches a tsc command in a script: at the start or after && or ;.
var plainTsc = regexp.MustCompile(`(^|&&\s*|;\s*)tsc(\s|$)`)

// isTestPattern is true for an exclude entry that leaves out test files.
func isTestPattern(p string) bool {
	p = strings.ToLower(filepath.ToSlash(p))
	if strings.Contains(p, ".test.") || strings.Contains(p, ".spec.") || strings.Contains(p, "__tests__") {
		return true
	}
	for _, part := range strings.Split(p, "/") {
		if part == "test" || part == "tests" {
			return true
		}
	}
	return false
}

// includeTests makes the project's tsconfig.json cover its tests when it
// excludes them, because typed ESLint (the project service) lints each file
// with the tsconfig.json nearest to it and fails on a file that it leaves
// out. The file as it was becomes tsconfig.build.json, which the build keeps
// using (a plain tsc in a script becomes tsc -p tsconfig.build.json), and
// tsconfig.json extends it without the test exclusions. It leaves the
// project alone when there is no tsconfig.json, it excludes no tests, or
// tsconfig.build.json exists already.
//
// Parameters: dir is the project's folder.
// Returns: the note for the person, or ""; and an error from the file system.
func includeTests(dir string) (string, error) {
	path := filepath.Join(dir, "tsconfig.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", nil
	}
	if _, err := os.Stat(filepath.Join(dir, buildTSConfig)); err == nil {
		return "", nil
	}
	var cfg struct {
		Exclude []string `json:"exclude"`
	}
	if json.Unmarshal(jsondoc.StripComments(raw), &cfg) != nil {
		return "", nil
	}
	kept := []string{}
	for _, p := range cfg.Exclude {
		if !isTestPattern(p) {
			kept = append(kept, p)
		}
	}
	if len(kept) == len(cfg.Exclude) {
		return "", nil
	}
	project, _ := jsondoc.ParseDocument([]byte("{}"))
	project.SetString("extends", "./"+buildTSConfig)
	if err := project.SetJSON("exclude", kept); err != nil {
		return "", err
	}
	scripts, err := pointScriptsAtBuild(dir)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, buildTSConfig), raw, 0o644); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, project.Bytes(), 0o644); err != nil {
		return "", err
	}
	note := "tsconfig.json left out the tests, so typed ESLint could not lint them: the build configuration is now " + buildTSConfig + ", and tsconfig.json extends it with the tests"
	if len(scripts) > 0 {
		note += "; the " + strings.Join(scripts, ", ") + " script runs tsc -p " + buildTSConfig
	}
	return note + ".", nil
}

// pointScriptsAtBuild makes each package.json script that runs a plain tsc
// (no -p, --project or --build) run tsc -p tsconfig.build.json.
//
// Parameters: dir is the project's folder.
// Returns: the names of the scripts changed, and an error from the file system.
func pointScriptsAtBuild(dir string) ([]string, error) {
	path := filepath.Join(dir, "package.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return nil, nil
	}
	scripts := doc.Object("scripts", false)
	if scripts == nil {
		return nil, nil
	}
	var changed []string
	for _, name := range scripts.Keys() {
		cmd, ok := scripts.String(name)
		if !ok || !plainTsc.MatchString(cmd) || hasProjectFlag(cmd) {
			continue
		}
		scripts.SetString(name, plainTsc.ReplaceAllString(cmd, "${1}tsc -p "+buildTSConfig+"${2}"))
		changed = append(changed, name)
	}
	if len(changed) == 0 {
		return nil, nil
	}
	return changed, os.WriteFile(path, doc.Bytes(), 0o644)
}

// hasProjectFlag is true when a script already names the tsconfig tsc reads.
func hasProjectFlag(cmd string) bool {
	for _, f := range strings.Fields(cmd) {
		if f == "-p" || f == "--project" || f == "-b" || f == "--build" || strings.HasPrefix(f, "--project=") {
			return true
		}
	}
	return false
}
