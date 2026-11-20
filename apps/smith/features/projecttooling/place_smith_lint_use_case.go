package projecttooling

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/removalplan"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/eslintpeer"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// rootFlatConfigs are the names ESLint reads a root flat config from.
var rootFlatConfigs = []string{"eslint.config.mjs", "eslint.config.js", "eslint.config.cjs", "eslint.config.ts", "eslint.config.mts", "eslint.config.cts"}

// jsxA11yOverride is the npm override smith's ESLint stack writes:
// eslint-plugin-jsx-a11y (a dependency of the shared config) still names
// eslint 9 as its newest peer, and works with the eslint the config needs.
const jsxA11yOverride = "eslint-plugin-jsx-a11y"

// Outcome is what putting one of smith's tools in place did.
type Outcome struct {
	// Ready is true when smith's tool is in place, so the old one may go.
	Ready bool
	// Left are the steps smith could not take, each with its fix, in words.
	Left []string
}

// PlaceSmithLint puts smith's lint in place for the project at dir: the root
// eslint.config.mjs that loads @monosmith/eslint-config (a root flat config
// of the user's is made smith's, or the step stops with the reason), the
// config and an eslint it runs with in the root devDependencies, and a
// "lint": "eslint ." script in the project. The project's eslintConfig key
// (an ESLint 8 configuration) goes: the root config replaces it. It removes
// no package and no .eslintrc file: the removal plan does that once this
// step is Ready.
//
// Parameters: root is the workspace folder (absolute); dir the project folder,
// workspace-relative; out gets one line for each change.
// Returns: the outcome; an error for a file that cannot be read or written.
func PlaceSmithLint(root, dir string, out io.Writer) (Outcome, error) {
	cfg := ""
	for _, name := range rootFlatConfigs {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			cfg = name
			break
		}
	}
	if cfg == "" {
		if err := os.WriteFile(filepath.Join(root, "eslint.config.mjs"), []byte(removalplan.SmithLintConfig), 0o644); err != nil {
			return Outcome{}, err
		}
		fmt.Fprintln(out, "  lint: wrote eslint.config.mjs (smith's config, "+removalplan.LintConfigPackage+")")
	} else {
		raw, err := os.ReadFile(filepath.Join(root, cfg))
		if err != nil {
			return Outcome{}, err
		}
		next, _, why, ok := removalplan.SwapLintConfig(string(raw))
		if !ok {
			return Outcome{Left: []string{"the root " + cfg + " passes options smith's config does not take (" + why + "). Make it the two lines of smith's config, then remove the old lint packages"}}, nil
		}
		if next != string(raw) {
			if err := os.WriteFile(filepath.Join(root, cfg), []byte(next), 0o644); err != nil {
				return Outcome{}, err
			}
			fmt.Fprintf(out, "  lint: %s now loads smith's config (%s)\n", cfg, removalplan.LintConfigPackage)
		}
	}
	var left []string
	err := editManifest(filepath.Join(root, "package.json"), func(doc *jsondoc.Document) error {
		text, err := removalplan.AddDevDependency(string(doc.Bytes()), removalplan.LintConfigPackage, removalplan.LintConfigRange)
		if err != nil {
			return err
		}
		if text, err = removalplan.AddDevDependency(text, "eslint", eslintpeer.Range); err != nil {
			return err
		}
		text, _, problem, err := removalplan.RaiseESLint(text)
		if err != nil {
			return err
		}
		if problem != "" {
			left = append(left, "package.json: "+problem+". Set eslint to "+eslintpeer.Range+", or ESLint cannot load smith's config")
		}
		next, err := jsondoc.ParseDocument([]byte(text))
		if err != nil {
			return err
		}
		*doc = *next
		overrides := doc.Object("overrides", true)
		if !overrides.Has(jsxA11yOverride) {
			overrides.Object(jsxA11yOverride, true).SetString("eslint", "$eslint")
		}
		return nil
	})
	if err != nil {
		return Outcome{}, err
	}
	err = editManifest(filepath.Join(root, filepath.FromSlash(dir), "package.json"), func(doc *jsondoc.Document) error {
		scripts := doc.Object("scripts", true)
		if lint, ok := scripts.String("lint"); !ok || lint != "eslint ." && containsWord(lint, "eslint") {
			scripts.SetString("lint", "eslint .")
		}
		doc.Delete("eslintConfig")
		if dir == "." {
			return nil
		}
		text, _, problem, err := removalplan.RaiseESLint(string(doc.Bytes()))
		if err != nil {
			return err
		}
		if problem != "" {
			left = append(left, path.Join(dir, "package.json")+": "+problem)
		}
		next, err := jsondoc.ParseDocument([]byte(text))
		if err != nil {
			return err
		}
		*doc = *next
		return nil
	})
	if err != nil {
		return Outcome{}, err
	}
	fmt.Fprintf(out, "  lint: %s lints with \"eslint .\" and the root config\n", path.Join(dir, "package.json"))
	return Outcome{Ready: len(left) == 0, Left: left}, nil
}

// editManifest reads a package.json, lets edit change it, and writes it back
// when it changed. The key order and the layout of what stays are kept.
func editManifest(file string, edit func(doc *jsondoc.Document) error) error {
	raw, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}
	if err := edit(doc); err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}
	next := doc.Bytes()
	if string(next) == string(raw) {
		return nil
	}
	return os.WriteFile(file, next, 0o644)
}

// containsWord reports whether a script runs the program word.
func containsWord(script, word string) bool {
	words := strings.FieldsFunc(script, func(r rune) bool { return r == ' ' || r == '&' || r == '|' || r == ';' })
	return slices.Contains(words, word)
}
