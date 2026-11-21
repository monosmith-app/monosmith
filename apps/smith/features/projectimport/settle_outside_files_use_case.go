package projectimport

import (
	"bufio"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/features/outsidefiles"
	"github.com/monosmith-app/monosmith/apps/smith/shared/choiceprompt"
	"github.com/monosmith-app/monosmith/apps/smith/shared/problem"
)

// The answers of --outside-files.
const (
	// OutsideCopy copies each file into the project folder and writes its new
	// path in the file that names it.
	OutsideCopy = "copy"
	// OutsideLeave leaves the files in the source; Left to do names each.
	OutsideLeave = "leave"
)

// SettleOutside settles what happens to the files the project uses outside
// its --path sub-folder, before anything changes. --outside-files answers; in
// a terminal without --yes smith asks; otherwise, and in a dry run, the files
// stay in the source and the Left to do list names each. Nothing is asked
// when no file can come along.
//
// Parameters: p is the plan, whose Outside is filled and whose CopyOutside
// the answer sets; o are the options.
// Returns: a problem for an --outside-files value that is not a choice.
func SettleOutside(p *Plan, o Options) error {
	switch o.Outside {
	case "", OutsideCopy, OutsideLeave:
	default:
		return problem.New("--outside-files "+o.Outside+" is not a choice", "use --outside-files copy or --outside-files leave")
	}
	var movable []string
	for _, f := range p.Outside {
		if f.Movable() {
			movable = append(movable, f.Ref)
		}
	}
	if len(movable) == 0 {
		return nil
	}
	answer := o.Outside
	if answer == "" && !o.DryRun && o.In != nil && operationplan.ShouldAsk(o.Env, o.Yes) {
		in, ok := o.In.(*bufio.Reader)
		if !ok {
			in = bufio.NewReader(o.In)
		}
		answer = choiceprompt.Ask(
			p.Name+" uses "+strings.Join(movable, ", ")+", outside "+p.Source.Path+". The import takes "+p.Source.Path+" only, so the build would not find them.",
			[]choiceprompt.Option{
				{Label: "Copy", Value: OutsideCopy, Hint: "copy each file into " + p.Folder + " and write its new path in the file that names it"},
				{Label: "Leave", Value: OutsideLeave, Hint: "copy nothing; Left to do names each file"},
			}, OutsideCopy, in, o.Out)
	}
	p.CopyOutside = answer == OutsideCopy
	return nil
}

// OutsideLines are the lines of the summary about the files outside the
// sub-folder: one per file, with what happens to it.
//
// Parameters: p is the settled plan.
// Returns: the lines, without indentation; empty when there is no such file.
func OutsideLines(p Plan) []string {
	var lines []string
	for _, f := range p.Outside {
		switch {
		case f.Movable() && p.CopyOutside:
			lines = append(lines, fmt.Sprintf("%s (%s names it): copied into %s as %s", f.Ref, f.Manifest, p.Folder, f.Name()))
		case f.Movable():
			lines = append(lines, fmt.Sprintf("%s (%s names it): left in the source (--outside-files copy brings it)", f.Ref, f.Manifest))
		default:
			lines = append(lines, outsideProblem(f, p.Source.Path))
		}
	}
	return lines
}

// BringOutside copies the movable files into the project folder when the
// answer was copy, and returns an item left for each file the project still
// misses.
//
// Parameters: root is the workspace root; p is the plan; out gets the progress.
// Returns: the items left, one per file that is not in the project folder.
func BringOutside(root string, p Plan, out io.Writer) []Left {
	if len(p.Outside) == 0 {
		return nil
	}
	failed := map[string]error{}
	brought := map[string]bool{}
	if p.CopyOutside {
		var done []outsidefiles.File
		done, failed = outsidefiles.Bring(filepath.Join(root, filepath.FromSlash(p.Folder)), p.Outside)
		for _, f := range done {
			brought[f.Path] = true
			fmt.Fprintf(out, "  outside: copied %s into %s; %s names %s now\n", f.Path, p.Folder, f.Manifest, f.Name())
		}
	}
	var left []Left
	for _, f := range p.Outside {
		if brought[f.Path] {
			continue
		}
		left = append(left, outsideLeft(p, f, failed[f.Path]))
	}
	return left
}

// outsideLeft is the item left for one file the project misses.
func outsideLeft(p Plan, f outsidefiles.File, failed error) Left {
	where := p.Folder + "/" + f.Manifest
	switch {
	case f.Movable():
		problemText := p.Name + ": the build needs " + f.Path + " from the source (" + f.Manifest + " names " + f.Ref + "), and it is outside " + p.Source.Path
		if failed != nil {
			problemText += ". It was not copied: " + failed.Error()
		}
		return Left{Problem: problemText, Remedy: "copy " + f.Path + " from the source into " + p.Folder + ", then write " + f.Name() + " in place of " + f.Ref + " in " + where + "\nor import again with --outside-files copy"}
	case f.Kind == outsidefiles.KindProject:
		return Left{Problem: p.Name + ": " + outsideProblem(f, p.Source.Path), Remedy: "import that project too (smith import <source> --path " + path.Dir(f.Path) + "), then point the reference in " + where + " at its new folder\nor remove the reference"}
	case f.Kind == outsidefiles.KindAbove:
		return Left{Problem: p.Name + ": " + outsideProblem(f, p.Source.Path), Remedy: "copy the settings " + p.Name + " needs from " + f.Path + " in the source into its project file in " + p.Folder}
	}
	return Left{Problem: p.Name + ": " + outsideProblem(f, p.Source.Path), Remedy: "copy what the build needs into " + p.Folder + ", then change the path in " + where}
}

// outsideProblem says in words why a file cannot come along.
func outsideProblem(f outsidefiles.File, sub string) string {
	switch {
	case f.Kind == outsidefiles.KindAbove:
		return "MSBuild reads " + f.Path + " from a folder above " + sub + " in the source, and it does not come with the import"
	case f.Missing:
		return f.Manifest + " names " + f.Ref + ", outside " + sub + ", and the source does not have " + f.Path
	case f.Kind == outsidefiles.KindProject:
		return "the build needs the project " + f.Path + " (" + f.Manifest + " references " + f.Ref + "), outside " + sub
	}
	return f.Manifest + " names the folder " + f.Ref + ", outside " + sub + ", and a folder does not come along"
}
