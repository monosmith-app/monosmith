package taskrun

import (
	"path"

	"github.com/monosmith-app/monosmith/apps/smith/marks"
	"github.com/monosmith-app/monosmith/apps/smith/shiplink"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// shipLines are the lines a run prints about what a task shipped into its
// project's shipped folder: the count, then each warning with its remedy. A
// warning never fails the task: the tool built correctly.
func shipLines(task string, rep shiplink.Report) []string {
	var lines []string
	if rep.Shipped > 0 {
		lines = append(lines, task+" │ "+words.Count("shiplink.shipped", rep.Shipped, "folder", rep.Folder))
	}
	if rep.Copied > 0 {
		lines = append(lines, task+" │ "+words.Count("shiplink.copied", rep.Copied, "folder", rep.Folder))
	}
	for _, f := range rep.InUse {
		program := path.Base(f)
		lines = append(lines,
			task+" "+marks.Warning+" "+words.Say("shiplink.in_use", "program", program, "folder", rep.Folder),
			task+" → "+words.Say("shiplink.in_use.remedy", "program", program))
	}
	for _, f := range rep.Failed {
		lines = append(lines, task+" "+marks.Warning+" "+words.Say("shiplink.failed", "file", f.File, "reason", f.Reason))
	}
	if rep.NothingMatched {
		lines = append(lines,
			task+" "+marks.Warning+" "+words.Say("shiplink.none", "folder", rep.Folder),
			task+" → "+words.Say("shiplink.none.remedy"))
	}
	return lines
}
