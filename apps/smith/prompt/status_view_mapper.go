package prompt

import (
	"fmt"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/marks"
)

// maxNamedProjects is how many affected projects the line names; the rest
// show as a count, so the line stays short in a prompt.
const maxNamedProjects = 3

// Doctor is the result of the last doctor run, as `smith prompt --json` shows it.
type Doctor struct {
	// Failed is the number of failed checks.
	Failed int `json:"failed"`
	// Warnings is the number of warnings.
	Warnings int `json:"warnings"`
	// CheckedAt is when doctor last ran.
	CheckedAt time.Time `json:"checkedAt"`
}

// View is everything `smith prompt` knows. Projects, Doctor and Affected are
// absent until a smith doctor run has saved a summary.
type View struct {
	// Version is the smith version.
	Version string `json:"version"`
	// Projects is the number of projects, from the last doctor run.
	Projects *int `json:"projects,omitempty"`
	// Doctor is the result of the last doctor run.
	Doctor *Doctor `json:"doctor,omitempty"`
	// Affected are the projects the working tree changes touch, with their dependents.
	Affected []string `json:"affected,omitempty"`
}

// NewView joins the version, the saved summary (ok is false when there is
// none) and the changed files of the working tree.
func NewView(version string, summary Summary, ok bool, changed []string) View {
	v := View{Version: version}
	if !ok {
		return v
	}
	n := len(summary.Projects)
	v.Projects = &n
	v.Doctor = &Doctor{Failed: summary.Failed, Warnings: summary.Warnings, CheckedAt: summary.CheckedAt}
	v.Affected = AffectedBy(summary.Projects, changed)
	return v
}

// Line is the one line a prompt prints: "smith 1.2.3 · 24 projects · × 2 ▲ 3 · api, web +2".
// Without a saved summary it is the version alone, and it never says anything
// is wrong. A clean last run shows the passed mark.
func (v View) Line() string {
	parts := []string{"smith " + v.Version}
	if v.Projects != nil {
		parts = append(parts, plural(*v.Projects, "project"))
	}
	if d := v.Doctor; d != nil {
		var marked []string
		if d.Failed > 0 {
			marked = append(marked, fmt.Sprintf("%s %d", marks.Failed, d.Failed))
		}
		if d.Warnings > 0 {
			marked = append(marked, fmt.Sprintf("%s %d", marks.Warning, d.Warnings))
		}
		if len(marked) == 0 {
			marked = append(marked, marks.Passed)
		}
		parts = append(parts, strings.Join(marked, " "))
	}
	if len(v.Affected) > 0 {
		shown := v.Affected
		rest := ""
		if len(shown) > maxNamedProjects {
			rest = fmt.Sprintf(" +%d", len(shown)-maxNamedProjects)
			shown = shown[:maxNamedProjects]
		}
		parts = append(parts, strings.Join(shown, ", ")+rest)
	}
	return strings.Join(parts, " · ")
}

func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
