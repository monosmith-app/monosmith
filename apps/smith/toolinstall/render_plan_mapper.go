package toolinstall

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/marks"
)

// RenderNeed words one tool that is missing or outdated: what it is, why smith
// wants it, which version.
func RenderNeed(st Status) string {
	var b strings.Builder
	if st.Exact && st.State == Outdated {
		fmt.Fprintf(&b, "\n%s %s.\n", st.Title, st.Problem())
		fmt.Fprintf(&b, "  Why smith needs it: %s.\n", st.Why)
		b.WriteString("  Keep yours if you prefer it: lint results can then differ from CI.\n")
		return b.String()
	}
	switch st.State {
	case Outdated:
		fmt.Fprintf(&b, "\n%s is too old: %s is installed.\n", st.Title, st.Version)
	default:
		fmt.Fprintf(&b, "\n%s is not installed.\n", st.Title)
	}
	fmt.Fprintf(&b, "  Why smith needs it: %s.\n", st.Why)
	if st.Wanted.Any() {
		return b.String()
	}
	fmt.Fprintf(&b, "  Version wanted: %s (from %s).\n", st.Wanted, st.WantedFrom)
	return b.String()
}

// RenderWays lists every way to get a tool: the install commands of the
// managers on the machine and the official page.
func RenderWays(plan ToolPlan) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Ways to get %s:\n", plan.Title)
	for _, o := range plan.Options {
		if o.Kind == OptionInstall {
			fmt.Fprintf(&b, "  %s\n", strings.Join(o.Command, " "))
		}
	}
	fmt.Fprintf(&b, "  or go to %s\n", plan.Download)
	b.WriteString("Then open a new terminal and run: smith tools\n")
	return b.String()
}

// RenderPlans words what `smith tools install --dry-run` shows: for each tool
// the options, the marked default and what each does.
func RenderPlans(plans []ToolPlan) string {
	var b strings.Builder
	for _, p := range plans {
		b.WriteString(RenderNeed(p.Status))
		for i, o := range p.Options {
			mark := ""
			if o.Value == p.Default {
				mark = " (default)"
			}
			fmt.Fprintf(&b, "  %d) %s%s: %s\n", i+1, o.Label, mark, o.Hint)
			if o.Note != "" {
				fmt.Fprintf(&b, "     Note: %s.\n", o.Note)
			}
		}
	}
	return b.String()
}

// RenderList words `smith tools`: one line per tool with its state, then what
// to run when something is missing.
func RenderList(list []Status) string {
	var b strings.Builder
	missing := false
	for _, st := range list {
		mark, detail := marks.Passed, strings.TrimSpace(st.Program+" "+st.Version)
		switch st.State {
		case Missing:
			mark, detail = marks.Failed, "not installed"
			if st.Optional {
				mark = "!"
			}
		case Outdated:
			mark, detail = "!", st.Version+", wanted "+st.Wanted.String()
			if st.Exact {
				detail = st.Version + ", CI pins " + st.Suggested
			}
		}
		if st.Lacking() {
			missing = true
		}
		fmt.Fprintf(&b, "%s %s: %s\n", mark, st.Title, detail)
		if st.Lacking() {
			fmt.Fprintf(&b, "    If you skip it: %s.\n", st.Consequence)
		}
	}
	if missing {
		b.WriteString("\nTo get what is missing: smith tools install\n")
	}
	return b.String()
}
