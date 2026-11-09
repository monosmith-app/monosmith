package cipipeline

import (
	"context"
	"io"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// setupRemedy is the way forward after a failed setup section. A blocking
// step stops the phase, so later toolchains did not run.
const setupRemedy = "→ the end of the step's output is above; fix the cause, then run smith ci again"

// runSetupSections runs the setup steps with one section per plugin, in the
// order the plugins gave them: a rule (a CI group on GitHub and Azure) with the
// plugin's name, its steps' own output, then one status line. Steps the plugins
// did not claim run in a last section named "setup". A blocking step that fails
// closes its section as failed and stops the phase, as runSteps does.
//
// Parameters: root is the workspace root; c holds the steps and which plugin
// gave them; sections receives the section lines and keeps the summary; out is
// where the steps' output goes.
// Returns: the outcome of every step that ran, and the error of the first
// blocking step that failed.
func runSetupSections(ctx context.Context, root string, c Contributions, sections *SectionLog, out io.Writer) ([]stepOutcome, error) {
	sections.Remedy = setupRemedy
	var outcomes []stepOutcome
	rest := c.Setup
	run := func(name string, projects int, steps []protocol.Step) error {
		sections.Begin(name, projects)
		done, err := runSteps(ctx, root, steps, nil, out)
		outcomes = append(outcomes, done...)
		sections.End(err)
		return err
	}
	for _, section := range c.SetupSections {
		n := min(section.Steps, len(rest))
		if err := run(section.Plugin, section.Projects, rest[:n]); err != nil {
			return outcomes, err
		}
		rest = rest[n:]
	}
	if len(rest) > 0 {
		return outcomes, run("setup", -1, rest)
	}
	return outcomes, nil
}
