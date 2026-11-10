package toolinstall

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/choiceprompt"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// Runner runs an install command. The real one starts the program with the
// terminal attached, so the manager's own questions and the system's own
// permission window work. Tests give a fake: no test runs a real install.
type Runner interface {
	Run(ctx context.Context, argv []string, out io.Writer) error
}

// EnvRunner is a Runner that can also run a command with more environment
// entries (GOBIN=<folder>), for an install option that has Env.
type EnvRunner interface {
	Runner
	RunEnv(ctx context.Context, argv, env []string, out io.Writer) error
}

// ExecRunner runs a command with the process's own terminal.
type ExecRunner struct{}

// RunEnv implements EnvRunner.
func (ExecRunner) RunEnv(ctx context.Context, argv, env []string, out io.Writer) error {
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, out, out
	cmd.Env = append(os.Environ(), env...)
	return cmd.Run()
}

// Run implements Runner.
func (ExecRunner) Run(ctx context.Context, argv []string, out io.Writer) error {
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, out, out
	return cmd.Run()
}

// Outcome is what happened to one tool.
type Outcome struct {
	Tool string `json:"tool"`
	// Choice is the Value of the option taken; "" when nothing was asked.
	Choice string `json:"choice"`
	// Ran is true when an install command ran.
	Ran bool `json:"ran"`
	// State is the state of the tool after the flow.
	State State `json:"state"`
	// Optional copies the tool's flag, so a caller can tell a failure from a warning.
	Optional bool `json:"optional,omitempty"`
}

// Flow presents the install choice of each missing tool and carries it out.
type Flow struct {
	In  *bufio.Reader
	Out io.Writer
	// Interactive is true when a person is at the terminal.
	Interactive bool
	// Yes takes the marked default of every choice. It never runs a command
	// that needs administrator rights.
	Yes bool
	// WithAdmin lets smith run commands that need administrator rights
	// without asking again; the system still asks for its own permission.
	WithAdmin bool
	// Manager picks the install option of this package manager for every tool.
	Manager string
	Runner  Runner
	// Detector re-detects a tool after its install.
	Detector Detector
}

// NoTerminalProblem is the refusal when nobody can answer and no flag does.
func NoTerminalProblem() error {
	return problem.New("smith cannot ask how to install the tools: there is no terminal",
		"name the answer: --yes takes the marked default (never for commands that need administrator rights), --manager <name> picks a package manager, --dry-run only shows the plan")
}

// Run presents each plan and carries out the answer.
//
// Parameters: plans are the tools that are missing or outdated.
// Returns: one outcome per plan, in order. It returns an error when nobody can
// be asked and no flag answers, when --manager names a manager that a tool has
// no way for, or when the context ends. A failed install is not an error: it
// is an outcome whose state is still missing.
func (f Flow) Run(ctx context.Context, plans []ToolPlan) ([]Outcome, error) {
	if len(plans) > 0 && !f.Interactive && !f.Yes && f.Manager == "" {
		return nil, NoTerminalProblem()
	}
	var out []Outcome
	for _, plan := range plans {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		o, err := f.one(ctx, plan)
		if err != nil {
			return out, err
		}
		out = append(out, o)
	}
	return out, nil
}

// one presents and carries out the choice for one tool.
func (f Flow) one(ctx context.Context, plan ToolPlan) (Outcome, error) {
	o := Outcome{Tool: plan.Name, State: plan.State, Optional: plan.Optional}
	fmt.Fprint(f.Out, RenderNeed(plan.Status))
	value, err := f.choose(plan)
	if err != nil {
		return o, err
	}
	o.Choice = value
	opt, _ := plan.Find(value)
	switch opt.Kind {
	case OptionInstall:
		return f.install(ctx, plan, opt, o), nil
	case OptionPage:
		fmt.Fprintf(f.Out, "Go to %s, install %s, then open a new terminal and run: smith tools\n", plan.Download, plan.Title)
	case OptionMyself:
		fmt.Fprint(f.Out, RenderWays(plan))
	case OptionSkip:
		fmt.Fprintf(f.Out, "Skipped. %s: %s.\n", plan.Title, plan.Consequence)
	}
	return o, nil
}

// choose reads the answer: a flag, the marked default, or the person's pick.
func (f Flow) choose(plan ToolPlan) (string, error) {
	if f.Manager != "" {
		value := "install:" + f.Manager
		if _, ok := plan.Find(value); !ok {
			return "", problem.New(plan.Title+" has no way to install with "+f.Manager+" on this machine",
				"choose one of: "+strings.Join(waysOf(plan), ", ")+" (or run smith tools install "+plan.Name+" without --manager)")
		}
		return value, nil
	}
	if f.Yes || !f.Interactive {
		return plan.Default, nil
	}
	opts := make([]choiceprompt.Option, len(plan.Options))
	for i, o := range plan.Options {
		label := o.Label
		if o.Value == plan.Default {
			label += " (default)"
		}
		opts[i] = choiceprompt.Option{Label: label, Value: o.Value, Hint: o.Hint}
	}
	return choiceprompt.Ask("How do you want to get "+plan.Title+"?", opts, plan.Default, f.In, f.Out), nil
}

// waysOf lists the managers a plan can install with.
func waysOf(plan ToolPlan) []string {
	var out []string
	for _, o := range plan.Options {
		if o.Kind == OptionInstall {
			out = append(out, o.Manager)
		}
	}
	if len(out) == 0 {
		return []string{"none found on this machine; use the official page"}
	}
	return out
}

// install runs one install option after the administrator question.
func (f Flow) install(ctx context.Context, plan ToolPlan, opt Option, o Outcome) Outcome {
	command := strings.Join(opt.Command, " ")
	if opt.Note != "" {
		fmt.Fprintln(f.Out, "Note: "+opt.Note+".")
	}
	if opt.Admin && !f.WithAdmin {
		if !f.Interactive || f.Yes || f.Manager != "" {
			fmt.Fprintf(f.Out, "This command needs administrator rights, so smith does not run it for you:\n  %s\nRun it yourself, or add --with-admin to let smith run it (the system asks for its own permission).\n", command)
			return o
		}
		answer := choiceprompt.Ask("This command needs administrator rights:\n  "+command, []choiceprompt.Option{
			{Label: "I will run it myself (default)", Value: "myself", Hint: "smith prints the command and goes on"},
			{Label: "Let smith run it", Value: "run", Hint: "the system asks for your permission or password itself"},
			{Label: "Skip it", Value: "skip", Hint: "go on without " + plan.Title + ": " + plan.Consequence},
		}, "myself", f.In, f.Out)
		switch answer {
		case "myself":
			fmt.Fprintf(f.Out, "Run this, then open a new terminal and run smith tools:\n  %s\n", command)
			return o
		case "skip":
			fmt.Fprintf(f.Out, "Skipped. %s: %s.\n", plan.Title, plan.Consequence)
			return o
		}
	}
	fmt.Fprintf(f.Out, "Running: %s\n", strings.TrimSpace(strings.Join(opt.Env, " ")+" "+command))
	o.Ran = true
	run := func() error { return f.Runner.Run(ctx, opt.Command, f.Out) }
	if len(opt.Env) > 0 {
		er, ok := f.Runner.(EnvRunner)
		if !ok {
			fmt.Fprintf(f.Out, "This runner cannot add %s to the command, so nothing ran.\n", strings.Join(opt.Env, " "))
			o.Ran = false
			return o
		}
		run = func() error { return er.RunEnv(ctx, opt.Command, opt.Env, f.Out) }
	}
	if err := run(); err != nil {
		fmt.Fprintf(f.Out, "The install did not finish: %v\nTry the official page, %s, or install it yourself.\n", err, plan.Download)
		return o
	}
	after := DetectAll(ctx, f.Detector, []Need{plan.Need})[0]
	o.State = after.State
	switch after.State {
	case Found:
		fmt.Fprintf(f.Out, "%s is installed (%s).\n", plan.Title, strings.TrimSpace(after.Program+" "+after.Version))
	case Outdated:
		fmt.Fprintf(f.Out, "The install finished, but this terminal still finds %s %s first, not the version wanted. Another copy is earlier on PATH than the folder the install wrote to.\n  Run smith doctor: it names both folders. Then move the new folder earlier in PATH, or remove the old copy yourself. smith does not edit PATH or delete files.\n", plan.Title, after.Version)
	default:
		fmt.Fprintf(f.Out, "The install finished, but %s is not found in this terminal yet. Open a new terminal (the PATH updates there), then run: smith tools\n", plan.Title)
	}
	return o
}
