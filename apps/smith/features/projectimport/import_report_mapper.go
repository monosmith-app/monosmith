package projectimport

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/features/projecttooling"
	"github.com/monosmith-app/monosmith/apps/smith/shared/marks"
)

// Summary is the text smith import prints before it asks: what it found,
// where it goes, what happens to the history and the steps, in order.
//
// Parameters: p is the settled plan.
// Returns: the text, ending in a newline.
func Summary(p Plan) string {
	var b strings.Builder
	from := p.Source.Arg
	fmt.Fprintf(&b, "Import %s from %s\n", p.Name, from)
	fmt.Fprintf(&b, "  found    a %s %s (%s)\n", p.Ecosystem, p.Kind, p.Manifest)
	fmt.Fprintf(&b, "  folder   %s  (--as and --into change it)\n", p.Folder)
	fmt.Fprintf(&b, "  history  %s\n", historyLine(p))
	if p.Source.Dirty && p.History != HistoryCopy {
		fmt.Fprintf(&b, "  ! %s has changes that are not committed: the import takes only %s (%s)\n", from, p.Source.Ref, p.Source.Commit)
	}
	split := "the development tools go to the workspace, README and LICENSE stay with the project"
	if len(p.Lockfiles) > 0 {
		split += ", " + strings.Join(p.Lockfiles, ", ") + " is dropped (the workspace has one lockfile), and the versions it locked go into the workspace lockfile"
	}
	steps := []string{
		"place    " + placeLine(p),
		"split    " + split,
		"join     add the project to its language's workspace",
		"tooling  " + toolingLine(p),
		"CI       write the pipelines again, so CI builds the project",
		"check    smith install -p " + p.Name + ", then smith doctor",
	}
	b.WriteString("Steps:\n")
	for i, s := range steps {
		fmt.Fprintf(&b, "  %d. %s\n", i+1, s)
	}
	if p.History == HistoryCopy {
		b.WriteString("smith commits nothing: --commit commits the import as one operation.\n")
	} else {
		fmt.Fprintf(&b, "%s in this repository. smith commits nothing else: --commit commits the rest as one operation.\n", commitsMade(p))
	}
	return b.String()
}

// Changes is the plan of paths that smith import presents for the
// confirmation: the new folder, the lockfiles it drops and the workspace
// manifest the split and the join edit.
//
// Parameters: p is the settled plan.
// Returns: the plan.
func Changes(p Plan) operationplan.Plan {
	plan := operationplan.Plan{Title: "import " + p.Name + " into " + p.Folder}
	plan.Add("Place", operationplan.Line{Action: operationplan.Create, Path: p.Folder, Note: placeLine(p)})
	for _, l := range p.Lockfiles {
		plan.Add("Split", operationplan.Line{Action: operationplan.Remove, Path: p.Folder + "/" + l, Note: "the workspace has one lockfile"})
	}
	if m := workspaceManifest(p.Ecosystem); m != "" {
		plan.Add("Split and join", operationplan.Line{Action: operationplan.Change, Path: m, Note: "the tools and the project's place in the workspace"})
	}
	return plan
}

// Outcome is the closing text: "✓ imported <name>", then one "Left to do"
// list, or the next step when nothing is left.
//
// Parameters: p is the plan; left are the items the import could not finish.
// Returns: the text, ending in a newline.
func Outcome(p Plan, left []Left) string {
	var b strings.Builder
	if len(left) == 0 {
		fmt.Fprintf(&b, "%s imported %s into %s, and smith doctor finds nothing to fix\n", marks.Passed, p.Name, p.Folder)
		fmt.Fprintf(&b, "  next: smith run build -p %s\n", p.Name)
		return b.String()
	}
	fmt.Fprintf(&b, "%s imported %s into %s. Left to do (%d):\n", marks.Passed, p.Name, p.Folder, len(left))
	for _, l := range left {
		fmt.Fprintf(&b, "  × %s\n", l.Problem)
		for i, line := range strings.Split(l.Remedy, "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			mark := "  "
			if i == 0 {
				mark = "→ "
			}
			fmt.Fprintf(&b, "    %s%s\n", mark, line)
		}
	}
	b.WriteString("  next: follow the steps above, then run smith doctor\n")
	return b.String()
}

// toolingLine says what happens to the project's build and lint tools.
func toolingLine(p Plan) string {
	switch p.Tooling {
	case string(projecttooling.Smith):
		line := "smith's build and ESLint config go in, then each group of old tools is shown and removed"
		if len(p.Legacy) > 0 {
			line += " (it uses " + strings.Join(p.Legacy, ", ") + ")"
		}
		return line
	case string(projecttooling.Keep):
		return "the project keeps its own tools (lint, test, build); --tooling smith uses smith's"
	}
	return "the project keeps its own tools (smith asks this question for Node projects only)"
}

func historyLine(p Plan) string {
	switch p.History {
	case HistoryKeep:
		return fmt.Sprintf("kept: the %d commit(s) of %s (%s) stay, and git log -- %s shows them (--history squash makes one commit)", p.Source.Commits, p.Source.Ref, p.Source.Commit, p.Folder)
	case HistorySquash:
		return fmt.Sprintf("squashed: one commit holds %s (%s), with the source in its message", p.Source.Ref, p.Source.Commit)
	}
	if p.Source.Git {
		return "not kept: the files are copied (--history keep keeps the commits)"
	}
	return "none: a plain folder is copied, without .git, dependencies and caches"
}

func placeLine(p Plan) string {
	if p.History == HistoryCopy {
		return "copy the files into " + p.Folder
	}
	if p.History == HistorySquash {
		return "git subtree add --prefix=" + p.Folder + " --squash <source> " + p.Source.Ref
	}
	return "merge the commits of " + p.Source.Ref + ", written again under " + p.Folder
}

func commitsMade(p Plan) string {
	if p.History == HistorySquash {
		return "git subtree add makes 2 commits (the squashed source and the merge)"
	}
	return fmt.Sprintf("git merge makes 1 merge commit and brings in the %d commit(s) of the source", p.Source.Commits)
}

// workspaceManifest is the workspace file the split and the join edit.
func workspaceManifest(ecosystem string) string {
	switch ecosystem {
	case "node":
		return "package.json"
	case "python":
		return "pyproject.toml"
	}
	return ""
}
