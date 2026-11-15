package release

import (
	"fmt"
	"slices"
	"strings"
)

// APIChange is how the public API of a package differs between its last
// release and now: the names that are gone and the names that are new.
type APIChange struct {
	Removed []string
	Added   []string
}

// DiffAPI compares two sorted lists of exported names.
func DiffAPI(before, after []string) APIChange {
	var c APIChange
	for _, name := range before {
		if !slices.Contains(after, name) {
			c.Removed = append(c.Removed, name)
		}
	}
	for _, name := range after {
		if !slices.Contains(before, name) {
			c.Added = append(c.Added, name)
		}
	}
	return c
}

// Needs is the smallest bump the change allows: a removed name breaks callers
// (major), a new name is a feature (minor), no difference asks for nothing.
// Before 1.0 a major is a minor, as Version.Bump does.
func (c APIChange) Needs(current Version) Level {
	switch {
	case len(c.Removed) > 0:
		if current.Major == 0 {
			return Minor
		}
		return Major
	case len(c.Added) > 0:
		return Minor
	}
	return None
}

// APIFinding says that the commits of a package ask for a smaller bump than
// the change of its public API needs.
type APIFinding struct {
	Project string
	// Planned is the bump the commits give, and Next the version it makes.
	Planned Level
	Next    string
	// Needs is the bump the API change asks for, and Proposed the version it
	// makes.
	Needs    Level
	Proposed string
	Change   APIChange
}

// CheckAPIBump returns a finding when the API change needs a bigger bump than
// planned. A package that was never released, or one whose version a
// directive forced, has nothing to compare and gives none.
func CheckAPIBump(item Item, change APIChange) (APIFinding, bool) {
	planned, ok := levelNamed(item.Level)
	current, err := ParseVersion(item.Current)
	if !ok || err != nil {
		return APIFinding{}, false
	}
	needs := change.Needs(current)
	if needs <= planned {
		return APIFinding{}, false
	}
	return APIFinding{
		Project: item.Project, Planned: planned, Next: item.Next,
		Needs: needs, Proposed: current.Bump(needs).String(), Change: change,
	}, true
}

// levelNamed is the Level a plan item names; "forced" and unknown names are
// not levels.
func levelNamed(name string) (Level, bool) {
	for l := None; l <= Major; l++ {
		if l.String() == name {
			return l, true
		}
	}
	return None, false
}

// maxNamesShown is how many names a warning lists before it says "and N more".
const maxNamesShown = 5

// Sentence is the warning, with the bump to use instead.
func (f APIFinding) Sentence() string {
	var what []string
	if n := len(f.Change.Removed); n > 0 {
		what = append(what, fmt.Sprintf("lost %s (%s)", count(n, "export"), listNames(f.Change.Removed)))
	}
	if n := len(f.Change.Added); n > 0 {
		what = append(what, fmt.Sprintf("gained %s (%s)", count(n, "export"), listNames(f.Change.Added)))
	}
	fix := "Commit it as `feat:` or `feat!:`, or take the new export out."
	if len(f.Change.Removed) > 0 {
		fix = "Mark a commit as breaking (`feat!:` or a `BREAKING CHANGE:` footer), or restore the export."
	}
	return fmt.Sprintf("%s: the commits release a %s (%s), but the public API %s; that needs a %s (%s). %s",
		f.Project, f.Planned, f.Next, strings.Join(what, " and "), f.Needs, f.Proposed, fix)
}

func count(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

func listNames(names []string) string {
	if len(names) <= maxNamesShown {
		return strings.Join(names, ", ")
	}
	return fmt.Sprintf("%s, and %d more", strings.Join(names[:maxNamesShown], ", "), len(names)-maxNamesShown)
}
