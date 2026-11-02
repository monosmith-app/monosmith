package doctorfix

import (
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/textdiff"
	"io"
	"os"
	"strings"
)

// shownSubjects is how many subjects a folded fix names before "and N more".
const shownSubjects = 4

// Scope says how many subjects a folded fix covers, "" for a single one.
func Scope(subjects []string) string {
	if len(subjects) < 2 {
		return ""
	}
	names := subjects
	more := ""
	if len(names) > shownSubjects {
		names, more = names[:shownSubjects], fmt.Sprintf(" and %d more", len(subjects)-shownSubjects)
	}
	return fmt.Sprintf("applies to %d projects: %s%s", len(subjects), strings.Join(names, ", "), more)
}

// ShowFix writes a fix as its description, the diff of each file, the
// commands it runs, and the questions it asks (the options of each).
func ShowFix(w io.Writer, root string, p Planned) error {
	fmt.Fprintf(w, "● %s\n", p.Description)
	if scope := Scope(p.AppliesTo); scope != "" {
		fmt.Fprintf(w, "  %s\n", scope)
	}
	for _, e := range p.Files {
		path, err := inside(root, e.Path)
		if err != nil {
			return err
		}
		before, _ := os.ReadFile(path)
		after := e.Content
		if e.Delete {
			after = ""
		}
		fmt.Fprint(w, textdiff.UnifiedDiff(e.Path, string(before), after))
	}
	for i, s := range p.Steps {
		if s.Description != "" {
			fmt.Fprintf(w, "  step %d: %s\n          $ %s\n", i+1, s.Description, displayCommand(s.Argv))
		} else {
			fmt.Fprintf(w, "  step %d: $ %s\n", i+1, displayCommand(s.Argv))
		}
	}
	for _, c := range p.Choices {
		fmt.Fprintf(w, "  asks: %s\n", c.Question)
		for i, o := range c.Options {
			fmt.Fprintf(w, "    %d) %s\n", i+1, o.Label)
		}
	}
	return nil
}
