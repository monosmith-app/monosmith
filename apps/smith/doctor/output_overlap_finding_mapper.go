package doctor

import (
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/doctorchecks"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/buildoutputs"
	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/smithfix"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// outputOverlapCheck is the name of the finding; doctor --fix --only names it.
const outputOverlapCheck = "no two tasks write the same output directory"

// outputOverlapFinding reports the tasks whose output folders overlap, in
// plain words: which tasks, why it goes wrong, and two fixes the person
// chooses from (ADR 0081). The fix is a question: give each task its own
// folder (smith shows the edit and the command, it changes neither), or tell
// smith that the tasks never run together ([workspace] alternative_outputs).
//
// Parameters: projects are the discovered projects; allowed are the folders
// tasks fill on purpose ([workspace] shared_outputs); alternatives are the
// groups of tasks that never run together; tomlSrc is the text of smith.toml,
// from which the second fix is made ("" offers no fix that edits it).
// Returns: the finding; it passes when no two tasks overlap.
func outputOverlapFinding(projects []*projectgraph.Project, allowed []string, alternatives [][]string, tomlSrc string) protocol.Finding {
	overlaps := buildoutputs.Overlaps(projects, allowed, alternatives)
	f := protocol.Finding{Check: outputOverlapCheck, OK: len(overlaps) == 0, Warning: true, Area: doctorchecks.AreaLint}
	if f.OK {
		return f
	}
	var lines []string
	for _, o := range overlaps {
		lines = append(lines, o.Line())
	}
	f.Problem = lines[0]
	if len(lines) > 1 {
		f.Problem = strconv.Itoa(len(lines)) + " output folders are filled by more than one task"
	}
	f.Detail = strings.Join(lines, "; ") + ". smith keeps each task's output in its cache. Restoring one task from the cache first " +
		"clears its output folder, which deletes the files the other task wrote there. A task that runs clears its folder first too"
	var remedy strings.Builder
	remedy.WriteString("choose one. (1) Give each task its own folder (suggested):")
	for _, o := range overlaps {
		remedy.WriteString("\n  " + ownFolderAdvice(projects, o))
	}
	remedy.WriteString("\n(2) These tasks never run together: tell smith, and it stops reporting them. In smith.toml:\n  " +
		alternativesLine(alternatives, overlaps) +
		"\n  Cost: smith still clears the folder before each task and a cache restore still replaces it, so running two of them in one run keeps only the last one's files.")
	remedy.WriteString("\n(A folder that several tasks fill on purpose, each with different files, goes under [workspace] shared_outputs instead: smith then never clears it, so old files stay.)")
	remedy.WriteString("\nRun `smith doctor --fix --only \"" + outputOverlapCheck + "\" --choose 1` or `--choose 2`.")
	f.Remedy = remedy.String()

	first := overlaps[0].B
	own := smithfix.Step("Show the command and outputs of "+first+", so you can change them", "explain", first)
	option1 := protocol.FixChoiceOption{
		Label: "give each task its own folder: smith shows the edit (nothing is changed by smith, the command is yours to change)",
		Fix:   protocol.Fix{Description: "Show the tasks to give their own folders", Steps: []protocol.FixStep{own}},
	}
	option2 := protocol.FixChoiceOption{Label: "these tasks never run together: tell smith (writes alternative_outputs in smith.toml)"}
	if tomlSrc != "" {
		option2.Fix = protocol.Fix{Description: "Record the tasks that never run together", Files: []protocol.FileEdit{{
			Path:    workspace.FileName,
			Content: configedit.SetKey(tomlSrc, "workspace", "alternative_outputs", alternativesValue(alternatives, overlaps)),
		}}}
	}
	f.Fix = &protocol.Fix{
		Description: "Settle the tasks that write the same output folder",
		Choices: []protocol.FixChoice{{
			Question: "Some tasks write the same output folder, and one deletes the files of the other. What do you want?",
			Options:  []protocol.FixChoiceOption{option1, option2},
		}},
	}
	if tomlSrc == "" {
		f.Fix.Choices[0].Options = f.Fix.Choices[0].Options[:1]
	}
	return f
}

// ownFolderAdvice is the exact edit that gives the later task of an overlap its
// own folder: the smith.toml lines, and the command to change. The new folder
// is the old one with the task's name added, so the two folders do not nest.
func ownFolderAdvice(projects []*projectgraph.Project, o buildoutputs.Overlap) string {
	name, task, _ := strings.Cut(o.B, ":")
	var p *projectgraph.Project
	for _, candidate := range projects {
		if candidate.Name == name {
			p = candidate
		}
	}
	if p == nil {
		return o.B + ": give it a folder of its own"
	}
	t := p.Tasks[task]
	var folders []string
	for _, out := range t.Outputs {
		folders = append(folders, configedit.Quote(out+"-"+strings.ReplaceAll(task, ":", "-")))
	}
	advice := fmt.Sprintf("%s: in smith.toml, [projects.%s.tasks.%s] outputs = [%s]", o.B, configedit.TableKey(name), configedit.TableKey(task), strings.Join(folders, ", "))
	switch {
	case len(t.Command) >= 3 && t.Command[0] == "npm" && t.Command[1] == "run":
		advice += fmt.Sprintf(", and change the %q script in %s/package.json so the build writes there", t.Command[2], p.Root)
	case len(t.Command) > 0:
		advice += fmt.Sprintf(", and make its command write there (it runs `%s`; override it with command = [...] in the same table)", strings.Join(t.Command, " "))
	}
	return advice
}

// alternativesLine is the smith.toml line that records the groups: the ones
// the workspace has, and one new group for each overlap.
func alternativesLine(existing [][]string, overlaps []buildoutputs.Overlap) string {
	return "[workspace] alternative_outputs = " + alternativesValue(existing, overlaps)
}

// alternativesValue is the TOML value of alternative_outputs after the overlaps
// are added as groups of two.
func alternativesValue(existing [][]string, overlaps []buildoutputs.Overlap) string {
	groups := append([][]string(nil), existing...)
	for _, o := range overlaps {
		groups = append(groups, []string{o.A, o.B})
	}
	parts := make([]string, len(groups))
	for i, g := range groups {
		quoted := make([]string, len(g))
		for j, t := range g {
			quoted[j] = configedit.Quote(t)
		}
		parts[i] = "[" + strings.Join(quoted, ", ") + "]"
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
