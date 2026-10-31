package doctorfix

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/lineendings"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// shownOutput is how many trailing lines of a failed step's output are shown.
const shownOutput = 25

// StepError is the step that stopped a fix.
type StepError struct {
	// Number is the step's place in the fix, from 1, of Total.
	Number, Total int
	Step          protocol.FixStep
	// Err says how it failed; Output is what it printed.
	Err    error
	Output string
}

func (e *StepError) Error() string {
	return fmt.Sprintf("step %d of %d failed (%s): %v", e.Number, e.Total, strings.Join(e.Step.Argv, " "), e.Err)
}

func (e *StepError) Unwrap() error { return e.Err }

// Apply writes a fix's files, then runs its steps in order and stops at the
// first that fails. A fix that still has Choices must be resolved first.
// Progress goes to out; a failed step's output is in the returned StepError.
func Apply(ctx context.Context, root string, fix protocol.Fix, out io.Writer) error {
	if len(fix.Choices) > 0 {
		return fmt.Errorf("%q has questions to answer first", fix.Description)
	}
	if err := ApplyFiles(root, fix); err != nil {
		return err
	}
	return RunSteps(ctx, root, fix.Steps, out)
}

// RunSteps runs commands from the workspace root (or the step's own Dir),
// capturing what each prints, and stops at the first failure.
func RunSteps(ctx context.Context, root string, steps []protocol.FixStep, out io.Writer) error {
	for i, step := range steps {
		dir, err := inside(root, orRoot(step.Dir))
		if err != nil {
			return &StepError{Number: i + 1, Total: len(steps), Step: step, Err: err}
		}
		fmt.Fprintf(out, "… %s\n", stepName(step))
		var captured bytes.Buffer
		code, err := taskrun.RunProcess(ctx, taskrun.Spec{Argv: step.Argv, Dir: dir, Env: taskrun.Environment(nil), Out: &captured})
		if err == nil && code != 0 {
			err = fmt.Errorf("exited with code %d", code)
		}
		if err != nil {
			return &StepError{Number: i + 1, Total: len(steps), Step: step, Err: err, Output: tail(captured.String(), shownOutput)}
		}
		fmt.Fprintf(out, "✔ %s\n", stepName(step))
	}
	return nil
}

func orRoot(dir string) string {
	if dir == "" {
		return "."
	}
	return dir
}

// stepName is what progress says of a step: its words, or the command.
func stepName(s protocol.FixStep) string {
	if s.Description != "" {
		return s.Description
	}
	return strings.Join(s.Argv, " ")
}

func tail(text string, lines int) string {
	all := strings.Split(strings.TrimRight(text, "\r\n"), "\n")
	if len(all) > lines {
		all = append([]string{"…"}, all[len(all)-lines:]...)
	}
	return strings.Join(all, "\n")
}

// ApplyFiles writes a fix's files, not its steps.
func ApplyFiles(root string, fix protocol.Fix) error {
	for _, e := range fix.Files {
		path, err := inside(root, e.Path)
		if err != nil {
			return err
		}
		if e.Delete {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(lineendings.Resolve(root, e.Path, e.Content)), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// inside resolves a fix's workspace-relative path, refusing one that leaves
// the workspace: a fix comes from a plugin.
func inside(root, rel string) (string, error) {
	path := filepath.Join(root, filepath.FromSlash(rel))
	back, err := filepath.Rel(root, path)
	if err != nil || filepath.IsAbs(rel) || back == ".." || strings.HasPrefix(back, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("a fix may only change files in the workspace, not %s", rel)
	}
	return path, nil
}
