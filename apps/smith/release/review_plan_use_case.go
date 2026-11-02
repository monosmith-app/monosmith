package release

import (
	"bufio"
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/choiceprompt"
)

// Review asks the person what to do with the plan: release all of it, change
// a bump, hold a package back, or cancel. It repeats until the person
// releases or cancels, and shows the plan again after each change.
//
// Parameters: in and out are the terminal; plan is what the commits asked
// for; show prints a plan; adj are the choices already made with flags.
// Returns: the plan to release and whether the person wants to go on (false
// on cancel). An error only when a flag choice cannot apply.
func Review(in *bufio.Reader, out io.Writer, plan Plan, show func(Plan), adj Adjustments) (Plan, bool, error) {
	current := plan
	if len(adj.Hold) > 0 || len(adj.Bump) > 0 {
		var err error
		if current, _, err = Adjust(plan, adj); err != nil {
			return plan, false, err
		}
	}
	for len(current.Items) > 0 {
		answer := choiceprompt.Ask("These packages release, each for the reason shown. What now?", []choiceprompt.Option{
			{Label: "Release all of them", Value: "go", Hint: "go to the final question (default)"},
			{Label: "Change a bump", Value: "bump", Hint: "choose patch, minor or major for one package"},
			{Label: "Hold a package back", Value: "hold", Hint: "it does not release this time and keeps its version"},
			{Label: "Cancel", Value: "cancel", Hint: "nothing is tagged or published"},
		}, "go", in, out)
		switch answer {
		case "go":
			return current, true, nil
		case "cancel":
			return current, false, nil
		}
		names := make([]choiceprompt.Option, 0, len(current.Items))
		for _, i := range current.Items {
			names = append(names, choiceprompt.Option{Label: i.Project, Value: i.Project, Hint: fmt.Sprintf("%s -> %s (%s)", orUnreleased(i.Current), i.Next, i.Reason)})
		}
		project := choiceprompt.Ask("Which package?", names, names[0].Value, in, out)
		trial := Adjustments{Hold: append([]string(nil), adj.Hold...), Bump: map[string]string{}}
		for k, v := range adj.Bump {
			trial.Bump[k] = v
		}
		if answer == "hold" {
			trial.Hold = append(trial.Hold, project)
		} else {
			trial.Bump[project] = choiceprompt.Ask("Which bump for "+project+"?", []choiceprompt.Option{
				{Label: "patch", Value: "patch", Hint: "bug fixes only"},
				{Label: "minor", Value: "minor", Hint: "new features, nothing breaks"},
				{Label: "major", Value: "major", Hint: "a breaking change"},
			}, "patch", in, out)
		}
		next, notes, err := Adjust(plan, trial)
		if err != nil {
			fmt.Fprintln(out, err)
			continue
		}
		adj, current = trial, next
		for _, n := range notes {
			fmt.Fprintln(out, "  "+n)
		}
		show(current)
	}
	return current, len(current.Items) > 0, nil
}
