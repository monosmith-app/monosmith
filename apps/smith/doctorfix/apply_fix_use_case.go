package doctorfix

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/marks"
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
	return fmt.Sprintf("step %d of %d failed (%s): %v", e.Number, e.Total, StepCommand(e.Step), e.Err)
}

func (e *StepError) Unwrap() error { return e.Err }

// Noter is an output that also wants the live detail of a fix: what it is
// writing and each line a command prints. The terminal interface shows the last
// one under its spinner. Apply and RunSteps tell it only when out is a Noter, so
// the command line keeps its short "… step" and "✓ step" lines.
type Noter interface {
	// Note receives one line of detail, without its line end.
	Note(line string)
}

// noteWriter splits what a command prints into lines for a Noter.
type noteWriter struct {
	to  Noter
	buf []byte
}

func (w *noteWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			return len(p), nil
		}
		if line := strings.TrimSpace(string(w.buf[:i])); line != "" {
			w.to.Note(line)
		}
		w.buf = w.buf[i+1:]
	}
}

// Apply writes a fix's files, then runs its steps in order and stops at the
// first that fails. A fix that still has Choices must be resolved first.
// Progress goes to out; a failed step's output is in the returned StepError.
func Apply(ctx context.Context, root string, fix protocol.Fix, out io.Writer) error {
	_, err := ApplyWithFindings(ctx, root, fix, out)
	return err
}

// Finding is an advisory step that failed: the fix went on after it.
type Finding struct {
	Step protocol.FixStep
	// Output is everything the step printed.
	Output string
}

// ApplyWithFindings is Apply, and it also returns the advisory steps that
// failed, in the order they ran.
//
// Parameters: root is the workspace, fix the resolved fix, out the progress.
// Returns the findings, and a *StepError for a step that stopped the fix.
func ApplyWithFindings(ctx context.Context, root string, fix protocol.Fix, out io.Writer) ([]Finding, error) {
	if len(fix.Choices) > 0 {
		return nil, fmt.Errorf("%q has questions to answer first", fix.Description)
	}
	if n, ok := out.(Noter); ok && len(fix.Files) > 0 {
		n.Note(filesNote(fix.Files))
	}
	if err := ApplyFiles(root, fix); err != nil {
		return nil, err
	}
	return runSteps(ctx, root, fix.Steps, out)
}

// RunSteps runs commands from the workspace root (or the step's own Dir),
// capturing what each prints, and stops at the first failure. An advisory
// step that fails does not stop it: its mark is a warning, what it printed
// follows, and the next step runs.
func RunSteps(ctx context.Context, root string, steps []protocol.FixStep, out io.Writer) error {
	_, err := runSteps(ctx, root, steps, out)
	return err
}

func runSteps(ctx context.Context, root string, steps []protocol.FixStep, out io.Writer) ([]Finding, error) {
	noter, _ := out.(Noter)
	var findings []Finding
	for i, step := range steps {
		dir, err := inside(root, orRoot(step.Dir))
		if err != nil {
			return findings, &StepError{Number: i + 1, Total: len(steps), Step: step, Err: err}
		}
		fmt.Fprintf(out, "… %s\n", stepName(step))
		var captured bytes.Buffer
		var sink io.Writer = &captured
		if noter != nil {
			sink = io.MultiWriter(&captured, &noteWriter{to: noter})
		}
		code, err := taskrun.RunProcess(ctx, taskrun.Spec{Argv: resolveProgram(step.Argv), Dir: dir, Env: taskrun.Environment(stepEnv(step)), Out: sink})
		if err == nil && code != 0 {
			err = fmt.Errorf("exited with code %d", code)
		}
		if err != nil && step.Advisory {
			fmt.Fprintf(out, marks.Warning+" %s\n", stepName(step))
			if text := strings.TrimRight(captured.String(), "\r\n"); strings.TrimSpace(text) != "" {
				fmt.Fprintln(out, tail(text, shownOutput))
			}
			findings = append(findings, Finding{Step: step, Output: captured.String()})
			continue
		}
		if err != nil {
			return findings, &StepError{Number: i + 1, Total: len(steps), Step: step, Err: err, Output: tail(captured.String(), shownOutput)}
		}
		fmt.Fprintf(out, marks.Passed+" %s\n", stepName(step))
	}
	return findings, nil
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
	return displayCommand(s.Argv)
}

func tail(text string, lines int) string {
	all := strings.Split(strings.TrimRight(text, "\r\n"), "\n")
	if len(all) > lines {
		all = append([]string{"…"}, all[len(all)-lines:]...)
	}
	return strings.Join(all, "\n")
}

// filesNote says which files a fix is writing, for a Noter.
func filesNote(files []protocol.FileEdit) string {
	if len(files) == 1 {
		return "Writing " + files[0].Path
	}
	return fmt.Sprintf("Writing %d files, from %s", len(files), files[0].Path)
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
