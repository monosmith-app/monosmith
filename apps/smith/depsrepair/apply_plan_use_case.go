package depsrepair

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/marks"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Action is what one step of a repair does.
type Action int

// The actions of a repair, in the order each tool does them.
const (
	// ActionDelete deletes a file or a folder (a link as a link).
	ActionDelete Action = iota
	// ActionRun runs a command of the plan.
	ActionRun
	// ActionInstall asks the plugin to install its projects again.
	ActionInstall
)

// Step is one step of a repair, as it runs and as progress names it.
type Step struct {
	// Tool is the tool the step belongs to.
	Tool Tool
	// Action is what the step does.
	Action Action
	// Path is the workspace-relative path an ActionDelete deletes.
	Path string
	// Command is the command an ActionRun runs.
	Command protocol.RepairCommand
	// Name is the step as a person reads it ("node: delete node_modules").
	Name string
}

// Ports are what a repair changes the world with; tests replace them.
type Ports struct {
	// Delete deletes the file or folder at an absolute path, a link as a link.
	Delete func(ctx context.Context, path string) error
	// Run runs one command from the workspace root (or its Dir), writes "… name"
	// as it starts and "✓ name" when it ends, and returns a
	// *doctorfix.StepError with what it printed when it fails.
	Run func(ctx context.Context, root string, step protocol.FixStep, out io.Writer) error
	// Install asks the tool's plugin to install its projects again.
	Install func(ctx context.Context, tool Tool, out io.Writer) error
}

// Steps lists the steps a repair kind takes: for each tool that has the kind,
// its deletions, then its commands, then its install.
//
// Parameters: offer is the workspace's offer; kind is the repair kind.
//
// Returns: the steps in the order Apply runs them; none when no tool has the kind.
func Steps(offer Offer, kind protocol.RepairKind) []Step {
	var out []Step
	for _, t := range offer.Tools {
		plan, ok := t.Plan(kind)
		if !ok {
			continue
		}
		for _, p := range plan.Delete {
			out = append(out, Step{Tool: t, Action: ActionDelete, Path: p, Name: t.Plugin + ": delete " + p})
		}
		for _, c := range plan.Commands {
			name := t.Plugin + ": run " + strings.Join(c.Argv, " ")
			if c.Dir != "" {
				name += " in " + c.Dir
			}
			out = append(out, Step{Tool: t, Action: ActionRun, Command: c, Name: name})
		}
		if plan.Install {
			out = append(out, Step{Tool: t, Action: ActionInstall, Name: t.Plugin + ": " + installName(len(t.Projects))})
		}
	}
	return out
}

// installName says what an install step does.
func installName(projects int) string {
	if projects == 1 {
		return "install 1 project again"
	}
	return fmt.Sprintf("install %d projects again", projects)
}

// Apply carries out a repair kind: each step of Steps in order, and it stops
// at the first that fails. It writes "… name" to out as a step starts and
// "✓ name" when it ends, so a terminal and the TUI show the progress. Each
// path is checked again just before it is deleted (checkDeletePath), because
// the disk may have changed since the plan was shown.
//
// Parameters: ctx stops the repair; offer is the workspace's offer; kind is the
// repair the person picked; ports change the world; out receives the progress.
//
// Errors: an error when no tool has the kind, or a *doctorfix.StepError that
// names the step that failed, its number and what it printed.
func Apply(ctx context.Context, offer Offer, kind protocol.RepairKind, ports Ports, out io.Writer) error {
	steps := Steps(offer, kind)
	if len(steps) == 0 {
		return fmt.Errorf("no tool of the workspace has the repair %q", kind)
	}
	for i, st := range steps {
		var err error
		output := ""
		switch st.Action {
		case ActionDelete:
			fmt.Fprintf(out, "… %s\n", st.Name)
			var abs string
			if abs, err = checkDeletePath(offer.Root, st.Path, offer.ProjectRoots); err == nil {
				err = ports.Delete(ctx, abs)
			}
		case ActionRun:
			err = ports.Run(ctx, offer.Root, protocol.FixStep{Description: st.Name, Argv: st.Command.Argv, Dir: st.Command.Dir}, out)
			var se *doctorfix.StepError
			if errors.As(err, &se) {
				err, output = se.Err, se.Output
			}
		case ActionInstall:
			fmt.Fprintf(out, "… %s\n", st.Name)
			err = ports.Install(ctx, st.Tool, out)
		}
		if err != nil {
			argv := st.Command.Argv
			if st.Action != ActionRun {
				// The step is not a command: the error names it by its words.
				argv = []string{st.Name}
			}
			return &doctorfix.StepError{Number: i + 1, Total: len(steps), Step: protocol.FixStep{Description: st.Name, Argv: argv}, Err: err, Output: output}
		}
		if st.Action != ActionRun {
			fmt.Fprintf(out, marks.Passed+" %s\n", st.Name)
		}
	}
	return nil
}
