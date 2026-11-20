package projectimport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/cipipeline"
	"github.com/monosmith-app/monosmith/apps/smith/features/dependencies"
	"github.com/monosmith-app/monosmith/apps/smith/features/derive"
	"github.com/monosmith-app/monosmith/apps/smith/features/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/features/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/features/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/features/session"
	"github.com/monosmith-app/monosmith/apps/smith/features/taskrun"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspacejoin"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Left is one thing the import could not finish, with its first step.
type Left struct {
	// Problem says what is left, in one line.
	Problem string
	// Remedy is the step that clears it; more lines follow a newline.
	Remedy string
}

// selfPath is the smith that is running, for smith install.
var selfPath = os.Executable

// Join joins the placed project to its language's workspace through the
// workspace/join method of its plugin, writes the pipelines again so CI
// builds it, and records the derived files. A join fix that asks a question
// (which solution) is left to smith doctor --fix.
//
// Parameters: root is the workspace root; version is smith's; p is the plan;
// out gets the progress.
// Returns: the items left, or an error when the workspace cannot be read.
func Join(root, version string, p Plan, out io.Writer) ([]Left, error) {
	s, err := session.Open(session.Options{Dir: root, Version: version, Log: io.Discard})
	if err != nil {
		return nil, err
	}
	defer s.Close()
	project := byFolder(s.Graph, p.Folder)
	if project == nil {
		return []Left{{
			Problem: "smith does not see " + p.Folder + " as a project after the import",
			Remedy:  "check that " + p.Manifest + " is at the top of " + p.Folder + ", then run smith projects",
		}}, nil
	}
	var left []Left
	for _, c := range s.Plugins.Clients {
		if c.Info.Plugin.Name != project.Plugin {
			continue
		}
		findings, err := workspacejoin.Ask(c, []protocol.Project{project.Project})
		if err != nil {
			left = append(left, Left{Problem: "the " + project.Plugin + " plugin did not say whether " + p.Name + " is joined: " + firstLine(err.Error()), Remedy: "smith doctor checks it again"})
		}
		for _, f := range findings {
			if f.OK {
				fmt.Fprintf(out, "  joined: %s is in its language's workspace already\n", p.Folder)
				continue
			}
			if f.Fix == nil || !f.Fix.Safe || len(f.Fix.Choices) > 0 {
				left = append(left, Left{Problem: p.Name + " " + f.Problem, Remedy: f.Remedy})
				continue
			}
			fmt.Fprintf(out, "… join: %s\n", f.Fix.Description)
			if err := doctorfix.Apply(context.Background(), root, *f.Fix, out); err != nil {
				left = append(left, Left{Problem: "joining " + p.Name + " failed: " + firstLine(err.Error()), Remedy: "smith doctor --fix tries the join again"})
			}
		}
	}
	if changed, _, err := cipipeline.Regenerate(s); err != nil {
		left = append(left, Left{Problem: "the pipelines were not written again: " + firstLine(err.Error()), Remedy: "smith ci generate"})
	} else if len(changed) > 0 {
		fmt.Fprintf(out, "  CI: wrote %d pipeline file(s) again, so CI builds %s\n", len(changed), p.Name)
	}
	if _, err := derive.Sync(s); err != nil {
		left = append(left, Left{Problem: "the derived files were not recorded: " + firstLine(err.Error()), Remedy: "smith derive"})
	}
	return left, nil
}

// Install installs the project's dependencies as smith install -p <name>
// does, and shows the output.
//
// Parameters: root is the workspace root; version is smith's; name is the
// project; out gets the output.
// Returns: an item left when the install fails, else nil.
func Install(root, version, name string, out io.Writer) []Left {
	failed := []Left{{Problem: "smith install -p " + name + " failed (its output is above)", Remedy: "fix the cause it names, then run smith install -p " + name}}
	s, err := session.Open(session.Options{Dir: root, Version: version, Log: io.Discard})
	if err != nil {
		return failed
	}
	defer s.Close()
	fmt.Fprintf(out, "… smith install -p %s\n", name)
	if err := dependencies.Install(s, []string{name}, nil, false, out, dependencies.LinkOptions{}); err != nil {
		fmt.Fprintf(out, "  %s\n", firstLine(err.Error()))
		return failed
	}
	return nil
}

// Verify runs the project's build and test tasks with the smith that is
// running. Their output is shown only when one fails: the last lines, where
// smith names the failed task.
//
// Parameters: root is the workspace root; name is the project; out gets the
// output of a failure.
// Returns: an item left when a task fails, else nil.
func Verify(root, name string, out io.Writer) []Left {
	self, err := selfPath()
	if err != nil {
		return nil
	}
	fmt.Fprintf(out, "… smith run build test -p %s\n", name)
	var captured bytes.Buffer
	code, err := taskrun.RunProcess(context.Background(), taskrun.Spec{Argv: []string{self, "run", "build", "test", "-p", name}, Dir: root, Env: taskrun.Environment(nil), Out: &captured})
	if err == nil && code == 0 {
		fmt.Fprintf(out, "  %s\n", lastLine(captured.String()))
		return nil
	}
	var lines []string
	for _, line := range strings.Split(strings.TrimRight(captured.String(), "\r\n"), "\n") {
		// npm repeats the failed command in its own error lines; the cause
		// is in the lines before them.
		if !strings.Contains(line, "npm error") && !strings.Contains(line, "npm ERR!") {
			lines = append(lines, line)
		}
	}
	if len(lines) > shownLines {
		lines = lines[len(lines)-shownLines:]
	}
	fmt.Fprintln(out, strings.Join(lines, "\n"))
	return []Left{{Problem: "the build or the tests of " + name + " fail (the last lines are above)", Remedy: "smith run build test -p " + name + " shows the whole output; fix the cause it names"}}
}

// shownLines is how many lines of a failed build or test Verify shows.
const shownLines = 15

func lastLine(text string) string {
	lines := strings.Split(strings.TrimRight(text, "\r\n"), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}

// Check runs smith doctor on the workspace and returns each failed check as
// an item left.
//
// Parameters: root is the workspace root; version is smith's.
// Returns: the items left, or an error when the workspace cannot be read.
func Check(root, version string) ([]Left, error) {
	s, err := session.Open(session.Options{Dir: root, Version: version, Log: io.Discard})
	if err != nil {
		return nil, err
	}
	defer s.Close()
	var left []Left
	for _, f := range doctor.Run(s).Findings {
		if f.OK || f.Warning {
			continue
		}
		problem := f.Detail
		if f.Project != "" && f.Problem != "" {
			problem = f.Project + ": " + f.Problem
		}
		left = append(left, Left{Problem: problem, Remedy: f.Remedy})
	}
	return left, nil
}

func byFolder(g *projectgraph.Graph, folder string) *projectgraph.Project {
	for _, p := range g.Projects {
		if p.Root == folder {
			return p
		}
	}
	return nil
}
