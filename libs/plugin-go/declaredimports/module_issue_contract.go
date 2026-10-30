package declaredimports

import "fmt"

// IssueKind says how go.mod and the imports disagree.
type IssueKind string

const (
	// Missing is an imported package that no requirement provides.
	Missing IssueKind = "missing"
	// Unused is a direct requirement that nothing imports.
	Unused IssueKind = "unused"
	// MarkedIndirect is a requirement marked "// indirect" that the module
	// imports itself.
	MarkedIndirect IssueKind = "indirect"
)

// Issue is one disagreement between a module's go.mod and its imports.
type Issue struct {
	Kind IssueKind
	// Path is the imported package (Missing, MarkedIndirect) or the
	// requirement's module path (Unused).
	Path string
	// File is the importing file, relative to the module; empty for Unused.
	File string
}

func (i Issue) String() string {
	switch i.Kind {
	case Missing:
		return fmt.Sprintf("%s imports %s, which no requirement in go.mod provides", i.File, i.Path)
	case MarkedIndirect:
		return fmt.Sprintf("%s imports %s, which go.mod marks as indirect", i.File, i.Path)
	default:
		return fmt.Sprintf("go.mod requires %s, but nothing imports it", i.Path)
	}
}

// Result is what a check found.
type Result struct {
	// Method says how it was checked, for the output's first line.
	Method string
	Issues []Issue
	// Diff is what `go mod tidy` would change, when tidy was the method.
	Diff string
}

// Clean is true when go.mod needs no change.
func (r Result) Clean() bool { return len(r.Issues) == 0 && r.Diff == "" }
