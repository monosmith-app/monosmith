package convert

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
)

// ReportOptions shape how much of a plan the report prints.
type ReportOptions struct {
	// DryRun words the heading as a plan and ends with how to apply it.
	DryRun bool
	// Strict counts work convert cannot do as blocking.
	Strict bool
	// Verbose names every project a finding applies to, instead of a few.
	Verbose bool
}

// detailSections are the findings that are information, not a request: the
// ones the reader can skim after the verdict. Everything that needs a human
// is in the verdict's list.
var detailSections = []struct {
	severity convertplan.Severity
	title    string
}{
	{convertplan.Changed, "Behaves differently"},
	{convertplan.Mapped, "Converted"},
	{convertplan.Dropped, "Dropped"},
}

// sourceLabel names the workspace a plan converts from.
func sourceLabel(plan *convertplan.Plan) string {
	switch {
	case plan.Source == convertplan.SourceTurborepo:
		return "Turborepo workspace"
	case plan.Legacy:
		return "Nx workspace (set up by an earlier version of this tool)"
	}
	return "Nx workspace"
}

// Report renders a plan for a terminal: the verdict and what needs a decision
// first, then the details grouped by what they say.
func Report(plan *convertplan.Plan, opts ReportOptions) string {
	var b strings.Builder
	verb := "Converting"
	if opts.DryRun {
		verb = "Conversion plan for"
	}
	fmt.Fprintf(&b, "%s %s %q\n\n", verb, sourceLabel(plan), plan.Config.Workspace.Name)
	b.WriteString(Verdict(plan, opts.Strict))
	b.WriteString(Details(plan, opts))
	if opts.DryRun && !plan.Blocked(opts.Strict) {
		b.WriteString("\n" + NextSteps(plan, true))
	}
	return b.String()
}

// Verdict opens the report: whether anything blocks, how many projects
// convert, and a numbered list of what the reader should look at before
// applying (decisions, work convert cannot do, and what it cannot convert).
// With strict, manual and unsupported findings block as well.
func Verdict(plan *convertplan.Plan, strict bool) string {
	var b strings.Builder
	blockers, checks := attention(plan)
	switch {
	case plan.Blocked(strict) && len(blockers) > 0:
		fmt.Fprintf(&b, "Cannot convert yet: %s.\n", plural(len(blockers), "problem blocks", "problems block")+" the conversion")
	case plan.Blocked(strict):
		b.WriteString("Cannot convert in strict mode: something needs manual work or is not supported.\n")
	default:
		fmt.Fprintf(&b, "Ready to convert: %s. Nothing blocks the conversion.\n", projectCount(plan))
	}
	n := 0
	list := func(title string, findings []convertplan.Finding) {
		if len(findings) == 0 {
			return
		}
		fmt.Fprintf(&b, "\n%s\n", title)
		for _, g := range groupFindings(findings, false) {
			n++
			fmt.Fprintf(&b, "  %d. %s\n", n, g.line())
			if more := g.projectLine(false); more != "" {
				fmt.Fprintf(&b, "       %s\n", more)
			}
		}
	}
	list("Fix first:", blockers)
	title := plural(len(groupFindings(checks, false)), "thing", "things") + " to check before you apply:"
	if len(blockers) > 0 {
		title = "Also check:"
	}
	list(title, checks)
	return b.String()
}

// attention splits the findings that need a person: blockers, then what to
// decide or do (decisions, manual work, unsupported projects).
func attention(plan *convertplan.Plan) (blockers, checks []convertplan.Finding) {
	for _, group := range []func(convertplan.Finding) bool{
		func(f convertplan.Finding) bool { return f.Decide },
		func(f convertplan.Finding) bool { return !f.Decide && f.Severity == convertplan.Manual },
		func(f convertplan.Finding) bool { return !f.Decide && f.Severity == convertplan.Unsupported },
	} {
		for _, f := range plan.Findings {
			if f.Severity != convertplan.Blocker && group(f) {
				checks = append(checks, f)
			}
		}
	}
	for _, f := range plan.Findings {
		if f.Severity == convertplan.Blocker {
			blockers = append(blockers, f)
		}
	}
	return blockers, checks
}

// Details lists what the conversion changes, grouped so that one message
// that applies to every project is printed once, then what it removes.
func Details(plan *convertplan.Plan, opts ReportOptions) string {
	var b strings.Builder
	findings := withoutShadowed(plan.Findings)
	for _, s := range detailSections {
		var in []convertplan.Finding
		for _, f := range findings {
			if f.Severity == s.severity && !f.Decide {
				in = append(in, f)
			}
		}
		if len(in) == 0 {
			continue
		}
		title := s.title
		if s.severity == convertplan.Dropped {
			title += " (" + dropReason(plan) + ")"
		}
		if s.severity == convertplan.Mapped && !opts.Verbose {
			title += ", nothing to do"
		}
		fmt.Fprintf(&b, "\n%s (%d):\n", title, len(in))
		for _, g := range groupFindings(in, true) {
			fmt.Fprintf(&b, "  - %s\n", g.line())
			if more := g.projectLine(opts.Verbose); more != "" {
				fmt.Fprintf(&b, "      %s\n", more)
			}
		}
	}
	if len(plan.Delete) > 0 {
		b.WriteString(removals(plan.Delete, opts.Verbose))
	}
	b.WriteString(keepSection(plan))
	return b.String()
}

// NextSteps says what to do after the conversion (or, for a dry run, how to
// apply it).
func NextSteps(plan *convertplan.Plan, dryRun bool) string {
	var b strings.Builder
	n := 0
	step := func(format string, args ...any) {
		n++
		fmt.Fprintf(&b, "  %d. %s\n", n, fmt.Sprintf(format, args...))
	}
	if dryRun {
		b.WriteString("Nothing changed (dry run). To apply:\n  → smith convert\n\nThen:\n")
	} else {
		b.WriteString("Next:\n")
	}
	step("git diff: review what changed (git status lists the new files; commit them with the rest)")
	step("smith doctor: check the workspace")
	step("smith run build: check that everything still builds")
	if plan.Source == convertplan.SourceTurborepo {
		step("smith i: install dependencies this way from now on, not with pnpm or yarn")
	}
	step("commit")
	return b.String()
}

// withoutShadowed drops a "converted" finding when a "behaves differently"
// one covers the same project and task: the reader gets one statement, the
// one that says what it became.
func withoutShadowed(findings []convertplan.Finding) []convertplan.Finding {
	changed := map[string]bool{}
	for _, f := range findings {
		if f.Severity == convertplan.Changed && f.Project != "" {
			changed[f.Project+"\x00"+f.Subject] = true
		}
	}
	out := make([]convertplan.Finding, 0, len(findings))
	for _, f := range findings {
		if f.Severity == convertplan.Mapped && f.Project != "" && changed[f.Project+"\x00"+f.Subject] {
			continue
		}
		out = append(out, f)
	}
	return out
}

// group is one message and the projects it applies to.
type group struct {
	subject, message string
	projects         []string
}

func (g group) line() string {
	if len(g.projects) == 1 {
		return g.projects[0] + " " + g.subject + ": " + g.message
	}
	return g.subject + ": " + g.message
}

// projectLine names the projects of a message shared by several: all of them
// when verbose, otherwise a few and a count.
func (g group) projectLine(verbose bool) string {
	if len(g.projects) < 2 {
		return ""
	}
	shown := g.projects
	rest := ""
	if limit := 5; !verbose && len(shown) > limit {
		rest = fmt.Sprintf(" and %d more", len(shown)-limit)
		shown = shown[:limit]
	}
	return fmt.Sprintf("%d projects: %s%s", len(g.projects), strings.Join(shown, ", "), rest)
}

// groupFindings merges findings that say the same thing about different
// projects (the project's name aside), keeping the order of first appearance
// sorted by what they are about.
func groupFindings(findings []convertplan.Finding, sorted bool) []group {
	index := map[string]int{}
	var groups []group
	for _, f := range findings {
		message := f.Message
		key := f.Subject + "\x00" + message
		if f.Project != "" {
			message = nameless(f.Message, f.Project)
			key = f.Subject + "\x00" + message
		}
		i, ok := index[key]
		if !ok {
			index[key] = len(groups)
			groups = append(groups, group{subject: f.Subject, message: f.Message})
			i = len(groups) - 1
		}
		if f.Project != "" {
			groups[i].projects = append(groups[i].projects, f.Project)
			if len(groups[i].projects) > 1 {
				groups[i].message = message
			}
		}
	}
	for i := range groups {
		sort.Strings(groups[i].projects)
	}
	if !sorted {
		return groups
	}
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].subject != groups[j].subject {
			return groups[i].subject < groups[j].subject
		}
		return groups[i].message < groups[j].message
	})
	return groups
}

// nameless replaces a project's name in a message with <project>, where it
// stands as a whole word, so the same message of two projects compares equal.
func nameless(message, project string) string {
	re := regexp.MustCompile(`(^|[^A-Za-z0-9_.-])` + regexp.QuoteMeta(project) + `($|[^A-Za-z0-9_-]|\.($|\s))`)
	return re.ReplaceAllStringFunc(message, func(m string) string {
		loc := re.FindStringSubmatchIndex(m)
		return m[loc[2]:loc[3]] + "<project>" + m[loc[4]:loc[5]]
	})
}

// removals lists the files and folders the conversion deletes: a file that
// goes from several folders is one line.
func removals(paths []string, verbose bool) string {
	byName := map[string][]string{}
	var names []string
	for _, p := range paths {
		name := path.Base(p)
		if _, ok := byName[name]; !ok {
			names = append(names, name)
		}
		byName[name] = append(byName[name], p)
	}
	sort.Strings(names)
	var b strings.Builder
	fmt.Fprintf(&b, "\nRemoves (%s):\n", plural(len(paths), "file or folder", "files and folders"))
	for _, name := range names {
		where := byName[name]
		sort.Strings(where)
		if len(where) == 1 {
			fmt.Fprintf(&b, "  - %s\n", where[0])
			continue
		}
		dirs := make([]string, len(where))
		for i, w := range where {
			dirs[i] = path.Dir(w)
		}
		shown, rest := dirs, ""
		if !verbose && len(shown) > 3 {
			rest = fmt.Sprintf(" and %d more", len(shown)-3)
			shown = shown[:3]
		}
		fmt.Fprintf(&b, "  - %s in %d folders: %s%s\n", name, len(where), strings.Join(shown, ", "), rest)
	}
	return b.String()
}

// projectCount reads "23 projects (22 packages, 1 app)".
func projectCount(plan *convertplan.Plan) string {
	kinds := map[string]int{}
	var order []string
	unsupported := 0
	for _, p := range plan.Projects {
		if !p.Supported {
			unsupported++
			continue
		}
		kind := string(p.Kind)
		if kinds[kind] == 0 {
			order = append(order, kind)
		}
		kinds[kind]++
	}
	sort.Strings(order)
	var parts []string
	total := 0
	for _, k := range order {
		total += kinds[k]
		parts = append(parts, plural(kinds[k], k, k+"s"))
	}
	text := plural(total, "project", "projects")
	if len(parts) > 0 {
		text += " (" + strings.Join(parts, ", ") + ")"
	}
	if unsupported > 0 {
		text += fmt.Sprintf(" and %d it cannot convert", unsupported)
	}
	return text
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
}

// dropReason says why what was dropped is not needed.
func dropReason(plan *convertplan.Plan) string {
	if plan.Source == convertplan.SourceTurborepo {
		return "Turborepo, pnpm and yarn only"
	}
	return "Nx-only"
}
