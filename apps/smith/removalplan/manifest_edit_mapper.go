package removalplan

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// removableSections are the package.json sections a removal takes a package
// out of; peerDependencies is a package's published contract, never edited.
var removableSections = []string{"dependencies", "devDependencies", "optionalDependencies"}

// RemovePackages returns a package.json without the named packages: out of
// its dependency sections, its npm overrides, and a section left empty.
func RemovePackages(text string, names []string) (string, error) {
	doc, err := jsondoc.ParseDocument([]byte(text))
	if err != nil {
		return "", err
	}
	for _, section := range removableSections {
		deps := doc.Object(section, false)
		if deps == nil {
			continue
		}
		for _, name := range names {
			deps.Delete(name)
		}
		if deps.Len() == 0 {
			doc.Delete(section)
		}
	}
	if overrides := doc.Object("overrides", false); overrides != nil {
		for _, name := range names {
			overrides.Delete(name)
		}
		if overrides.Len() == 0 {
			doc.Delete("overrides")
		}
	}
	return string(doc.Bytes()), nil
}

// simpleRange is one version with an optional ^, ~, >= or = before it: the
// ranges whose lowest version can be read.
var simpleRange = regexp.MustCompile(`^(?:\^|~|>=|=)?v?(\d+)(?:\.(\d+))?(?:\.(\d+))?$`)

// versionRange is anything npm reads as a version range rather than a tag,
// a path, a URL or an alias.
var versionRange = regexp.MustCompile(`^[\d^~<>=*xX\s|.-]+$`)

// rangeFloor is the lowest version a simple range accepts ("^9" is 9.0.0);
// ok is false for any other range.
func rangeFloor(r string) (string, bool) {
	m := simpleRange.FindStringSubmatch(strings.TrimSpace(r))
	if m == nil {
		return "", false
	}
	part := func(s string) string {
		if s == "" {
			return "0"
		}
		return s
	}
	return m[1] + "." + part(m[2]) + "." + part(m[3]), true
}

// RaiseESLint returns a package.json whose eslint is one smith's config runs
// with (LintESLintRange): in its dependency sections and in a top-level npm
// override that names a range. A "$eslint" override follows the dependency
// and stays. peerDependencies are a package's published contract and are
// never edited.
//
// Parameters: text is the package.json. Returns the new text, the ranges it
// replaced (in the order met), and problem: why an eslint could not be raised
// ("" when nothing stands in the way). err is a manifest that is not JSON.
func RaiseESLint(text string) (next string, raised []string, problem string, err error) {
	doc, err := jsondoc.ParseDocument([]byte(text))
	if err != nil {
		return "", nil, "", err
	}
	var problems []string
	raise := func(where *jsondoc.Document, label string) {
		if where == nil {
			return
		}
		current, ok := where.String("eslint")
		if !ok || strings.HasPrefix(current, "$") {
			return
		}
		if floor, simple := rangeFloor(current); simple && ESLintMeetsConfig(floor) {
			return
		}
		if !versionRange.MatchString(current) {
			problems = append(problems, fmt.Sprintf("its %s name eslint %q, which smith cannot raise to %s", label, current, LintESLintRange))
			return
		}
		where.SetString("eslint", LintESLintRange)
		raised = append(raised, current)
	}
	for _, section := range removableSections {
		raise(doc.Object(section, false), section)
	}
	raise(doc.Object("overrides", false), "overrides")
	if peers := doc.Object("peerDependencies", false); peers != nil {
		if current, ok := peers.String("eslint"); ok && !strings.Contains(current, "||") {
			if floor, simple := rangeFloor(current); simple && !ESLintMeetsConfig(floor) {
				problems = append(problems, fmt.Sprintf("its peerDependencies ask for eslint %q, and smith's config needs %s", current, LintESLintRange))
			}
		}
	}
	if len(raised) == 0 {
		return text, nil, strings.Join(problems, "; "), nil
	}
	return string(doc.Bytes()), raised, strings.Join(problems, "; "), nil
}

// AddDevDependency returns a package.json with the package in its
// devDependencies, unless it already depends on it in any section.
func AddDevDependency(text, name, versionRange string) (string, error) {
	doc, err := jsondoc.ParseDocument([]byte(text))
	if err != nil {
		return "", err
	}
	for _, section := range append([]string{"peerDependencies"}, removableSections...) {
		if deps := doc.Object(section, false); deps != nil && deps.Has(name) {
			return text, nil
		}
	}
	doc.Object("devDependencies", true).SetString(name, versionRange)
	return string(doc.Bytes()), nil
}
