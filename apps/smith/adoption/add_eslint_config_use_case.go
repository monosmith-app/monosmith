package adoption

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// eslintConfigPackage is the shared lint config smith's vertical-slice and
// formatting rules ship in.
const eslintConfigPackage = "@monosmith/eslint-config"

const eslintConfigVersion = "^1.0.0"

const eslintConfigFile = "import monosmith from '" + eslintConfigPackage + "'\nexport default monosmith({ workspaceRoot: import.meta.dirname })\n"

var rootESLintConfigs = []string{"eslint.config.mjs", "eslint.config.js", "eslint.config.cjs", "eslint.config.ts", "eslint.config.mts", "eslint.config.cts"}

// AddESLintConfig does the step adopt otherwise leaves in a paragraph: the
// two-line root eslint.config.mjs, and @monosmith/eslint-config in the root
// devDependencies (npm install, run by --install, then installs it). A root
// ESLint config of the user's is never replaced; it says so.
func AddESLintConfig(root string, out io.Writer) error {
	for _, name := range rootESLintConfigs {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			fmt.Fprintf(out, "  kept your %s; to use smith's lint rules make it two lines (see smith doctor)\n", name)
			return nil
		}
	}
	if err := os.WriteFile(filepath.Join(root, "eslint.config.mjs"), []byte(eslintConfigFile), 0o644); err != nil {
		return err
	}
	fmt.Fprintln(out, "  wrote eslint.config.mjs (two lines: "+eslintConfigPackage+")")
	manifest := filepath.Join(root, "package.json")
	raw, err := os.ReadFile(manifest)
	if err != nil {
		fmt.Fprintf(out, "  not done: no root package.json; npm i -D %s\n", eslintConfigPackage)
		return nil
	}
	updated, changed, err := withDevDependency(raw, eslintConfigPackage, eslintConfigVersion)
	if err != nil {
		fmt.Fprintf(out, "  not done: %v; run npm i -D %s\n", err, eslintConfigPackage)
		return nil
	}
	if changed {
		if err := os.WriteFile(manifest, updated, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(out, "  added %s to the root devDependencies\n", eslintConfigPackage)
	}
	return nil
}

var devDependenciesKey = regexp.MustCompile(`"devDependencies"\s*:\s*\{`)

// withDevDependency returns a package.json whose devDependencies name the
// package, and says whether it changed. A package already listed there is
// left alone; the rest of the manifest stays exactly as written.
func withDevDependency(raw []byte, name, version string) ([]byte, bool, error) {
	var probe struct {
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return raw, false, fmt.Errorf("package.json: %w", err)
	}
	if _, has := probe.DevDependencies[name]; has {
		return raw, false, nil
	}
	entry := fmt.Sprintf("%q: %q", name, version)
	var out string
	if loc := devDependenciesKey.FindIndex(raw); loc != nil && probe.DevDependencies != nil {
		rest := string(raw[loc[1]:])
		sep := ","
		if strings.HasPrefix(strings.TrimLeft(rest, " \t\r\n"), "}") {
			sep = ""
		}
		out = string(raw[:loc[1]]) + "\n    " + entry + sep + rest
		if sep == "" {
			out = string(raw[:loc[1]]) + "\n    " + entry + "\n  " + strings.TrimLeft(rest, " \t\r\n")
		}
	} else {
		body := bytes.TrimRight(raw, " \t\r\n")
		end := bytes.LastIndexByte(body, '}')
		if end < 0 {
			return raw, false, fmt.Errorf("package.json has no closing brace")
		}
		head := bytes.TrimRight(body[:end], " \t\r\n")
		sep := ",\n  "
		if bytes.HasSuffix(head, []byte("{")) {
			sep = "\n  "
		}
		out = string(head) + sep + `"devDependencies": {` + "\n    " + entry + "\n  }\n}\n"
	}
	if json.Unmarshal([]byte(out), &probe) != nil {
		return raw, false, fmt.Errorf("package.json: could not add the devDependency")
	}
	return []byte(out), true, nil
}
