package doctorfix

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Options say how a plan is walked.
type Options struct {
	Root string
	Out  io.Writer
	// In is where answers are read from; used only when Interactive.
	In          io.Reader
	Interactive bool
	// Yes applies the safe fixes without asking.
	Yes bool
	// Explicit is true when the user named the fixes (--only). With Yes it also
	// applies a named fix that asks, once every question has an answer.
	Explicit bool
	// Choose are answers given up front, in question order, from 1.
	Choose []int
	// Command spells the command line that applies p with the given answers,
	// for a fix that cannot be asked here.
	Command func(p Planned, answers []int) string

	// snapshot is what the plan's files held when the walk began.
	snapshot map[string]string
}

// Pending is a fix that was shown but not applied because nobody could decide.
type Pending struct {
	Fix Planned
	// Commands are the command lines that would apply it, one for each answer
	// to its next question (one when it has none).
	Commands []string
}

// Result is what a walk did.
type Result struct {
	Applied []string
	Pending []Pending
	// Skipped are fixes the user declined at a prompt.
	Skipped []string
	// Stale are fixes not applied because a file they write was changed after
	// the plan was made, by a fix applied earlier in the walk: planning again
	// would make them from the new content.
	Stale []string
}

// errDeclined ends a prompt without applying; it is not a failure.
var errDeclined = errors.New("declined")

// Walk shows the plan and applies it. Safe fixes without questions go first,
// together, after one confirmation (or Yes). Each fix that asks follows: in a
// terminal its questions are asked as numbered prompts and the result is
// confirmed; elsewhere it is shown with the command that applies it, and
// listed as pending. A failing step stops the walk with a *StepError.
func Walk(ctx context.Context, plan []Planned, o Options) (Result, error) {
	var res Result
	o.snapshot = snapshotFiles(o.Root, plan)
	reader := bufio.NewReader(o.In)
	if o.In == nil {
		reader = bufio.NewReader(strings.NewReader(""))
	}
	var safe, asked []Planned
	for _, p := range plan {
		if p.Ask {
			asked = append(asked, p)
		} else {
			safe = append(safe, p)
		}
	}
	for _, p := range safe {
		if err := ShowFix(o.Out, o.Root, p); err != nil {
			return res, err
		}
	}
	if len(safe) > 0 {
		apply := o.Yes
		if !apply {
			if !o.Interactive {
				fmt.Fprintln(o.Out, "\nPass --yes to apply these fixes.")
			} else {
				apply = confirm(reader, o.Out, fmt.Sprintf("\nApply %d fix(es)?", len(safe)))
				if !apply {
					fmt.Fprintln(o.Out, "Nothing applied.")
				}
			}
		}
		if apply {
			for _, p := range safe {
				if err := applyOne(ctx, o, p, &res); err != nil {
					return res, err
				}
			}
		}
	}
	for _, p := range asked {
		fmt.Fprintln(o.Out)
		if err := walkAsked(ctx, o, reader, p, &res); err != nil {
			return res, err
		}
	}
	return res, nil
}

func walkAsked(ctx context.Context, o Options, in *bufio.Reader, p Planned, res *Result) error {
	if stale(o, p, res) {
		return nil
	}
	// The fix is shown once: before its first open question, or, when the
	// answers given up front settle every question, as the resolved fix only.
	shown := false
	show := func() error {
		if shown {
			return nil
		}
		shown = true
		return ShowFix(o.Out, o.Root, p)
	}
	answers := append([]int(nil), o.Choose...)
	for {
		fix, question, err := Resolve(p.Fix, answers)
		if err != nil {
			return err
		}
		if question != nil {
			if err := show(); err != nil {
				return err
			}
			if !o.Interactive {
				res.Pending = append(res.Pending, pendingFor(o, p, answers, question))
				return nil
			}
			pick, err := pickOption(in, o.Out, *question)
			if err != nil {
				if errors.Is(err, errDeclined) {
					res.Skipped = append(res.Skipped, p.Description)
					return nil
				}
				return err
			}
			answers = append(answers, pick)
			continue
		}
		// Every question is answered: show what the answers add, then decide.
		if len(p.Choices) > 0 {
			chosen := Planned{Fix: fix, Checks: p.Checks}
			chosen.Description = "your choices: " + p.Description
			if err := ShowFix(o.Out, o.Root, chosen); err != nil {
				return err
			}
		} else if err := show(); err != nil {
			return err
		}
		switch {
		case o.Interactive:
			if !confirm(in, o.Out, "Apply this fix?") {
				fmt.Fprintln(o.Out, "Not applied.")
				res.Skipped = append(res.Skipped, p.Description)
				return nil
			}
		case !o.Yes || !o.Explicit:
			res.Pending = append(res.Pending, pendingFor(o, p, answers, nil))
			return nil
		}
		return applyOne(ctx, o, Planned{Fix: fix, Checks: p.Checks}, res)
	}
}

// pendingFor lists the commands that finish a fix nobody could decide: one for
// each option of the question still open, or the one command when none is.
func pendingFor(o Options, p Planned, answers []int, question *protocol.FixChoice) Pending {
	pending := Pending{Fix: p}
	if o.Command == nil {
		return pending
	}
	if question == nil {
		pending.Commands = []string{o.Command(p, answers)}
		return pending
	}
	for i, opt := range question.Options {
		pending.Commands = append(pending.Commands, o.Command(p, append(append([]int(nil), answers...), i+1))+"   # "+opt.Label)
	}
	return pending
}

func applyOne(ctx context.Context, o Options, p Planned, res *Result) error {
	if stale(o, p, res) {
		return nil
	}
	if err := Apply(ctx, o.Root, p.Fix, o.Out); err != nil {
		var step *StepError
		if errors.As(err, &step) && step.Output != "" {
			fmt.Fprintln(o.Out, step.Output)
		}
		return fmt.Errorf("%s: %w", p.Description, err)
	}
	res.Applied = append(res.Applied, p.Description)
	fmt.Fprintf(o.Out, "✔ %s\n", p.Description)
	return nil
}

// pickOption asks a question as a numbered list; 0 or an empty answer declines.
func pickOption(in *bufio.Reader, out io.Writer, q protocol.FixChoice) (int, error) {
	fmt.Fprintf(out, "\n%s\n", q.Question)
	for i, opt := range q.Options {
		fmt.Fprintf(out, "  %d) %s\n", i+1, opt.Label)
	}
	fmt.Fprint(out, "  0) skip this fix\nChoose: ")
	line, _ := in.ReadString('\n')
	line = strings.TrimSpace(line)
	n, err := strconv.Atoi(line)
	if line == "" || (err == nil && n == 0) {
		return 0, errDeclined
	}
	if err != nil || n < 1 || n > len(q.Options) {
		fmt.Fprintf(out, "%q is not one of the options; skipped.\n", line)
		return 0, errDeclined
	}
	return n, nil
}

func confirm(in *bufio.Reader, out io.Writer, question string) bool {
	fmt.Fprintf(out, "%s [y/N] ", question)
	answer, _ := in.ReadString('\n')
	a := strings.ToLower(strings.TrimSpace(answer))
	return a == "y" || a == "yes"
}

// stale reports, and records, a fix one of whose files changed since the walk
// began: it was made from the old content and would undo what changed it.
func stale(o Options, p Planned, res *Result) bool {
	for _, path := range filePaths(p.Fix) {
		abs, err := inside(o.Root, path)
		if err != nil {
			return false
		}
		now, _ := os.ReadFile(abs)
		if before, ok := o.snapshot[path]; ok && before != string(now) {
			fmt.Fprintf(o.Out, "~ %s: not applied, %s changed after the plan was made (by a fix applied just before); it is planned again from the new content\n", p.Description, path)
			res.Stale = append(res.Stale, p.Description)
			return true
		}
	}
	return false
}

// snapshotFiles reads the files the plan's fixes write, options included.
func snapshotFiles(root string, plan []Planned) map[string]string {
	out := map[string]string{}
	for _, p := range plan {
		for _, path := range filePaths(p.Fix) {
			if abs, err := inside(root, path); err == nil {
				raw, _ := os.ReadFile(abs)
				out[path] = string(raw)
			}
		}
	}
	return out
}

// filePaths are the files a fix may write, those of its choices' options too.
func filePaths(f protocol.Fix) []string {
	var paths []string
	for _, e := range f.Files {
		paths = append(paths, e.Path)
	}
	for _, c := range f.Choices {
		for _, o := range c.Options {
			paths = append(paths, filePaths(o.Fix)...)
		}
	}
	return paths
}
