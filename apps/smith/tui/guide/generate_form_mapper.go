package guide

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/clicontract"
)

// notAsked are the flags a form never offers: the output format for tools,
// help, and the confirmation the shell asks for itself.
var notAsked = []string{"json", "help", "yes"}

// FormFor builds the form that asks for a command's arguments and options,
// from the description `smith commands` gives, so a command has a form
// without anyone writing one: the required arguments are asked first, then
// the optional ones, the switches in one toggle list (with a choice of none)
// and each option that takes a value, all of which may be left empty. path is
// the command's words under smith (["deps", "up"]) and id the action it presents.
func FormFor(c clicontract.CommandDescription, path []string, id string) Action {
	a := Action{Title: c.Description, Hint: "smith " + strings.Join(path, " "), NeedsWorkspace: true, Presents: id}
	args := slices.Clone(c.Arguments)
	slices.SortStableFunc(args, func(x, y clicontract.ArgumentDescription) int { return rank(y.Required) - rank(x.Required) })
	for _, arg := range args {
		step := Step{Key: "arg:" + arg.Name, Prompt: arg.Name, Placeholder: arg.Name, Optional: !arg.Required}
		if arg.Variadic {
			step.Prompt += " (one or more, separated by spaces)"
		}
		if !arg.Required {
			step.Prompt += " (optional)"
		}
		for _, choice := range arg.Choices {
			step.Options = append(step.Options, Option{Label: choice, Value: choice})
		}
		a.Steps = append(a.Steps, step)
	}
	var switches []Option
	var valued []clicontract.OptionDescription
	for _, o := range c.Options {
		switch {
		case slices.Contains(notAsked, o.Name):
		case o.TakesValue == "none":
			switches = append(switches, Option{Label: "--" + o.Name, Value: "--" + o.Name, Hint: o.Description})
		default:
			valued = append(valued, o)
		}
	}
	if len(switches) > 0 {
		none := Option{Label: "none", Value: "", Hint: "run it as it is"}
		a.Steps = append(a.Steps, Step{Key: "switches", Prompt: "Options? (space toggles, enter confirms)", Multi: true, Options: append([]Option{none}, switches...)})
	}
	for _, o := range valued {
		prompt := "--" + o.Name + " (optional)"
		if o.Description != "" {
			prompt = "--" + o.Name + ": " + o.Description + " (optional)"
		}
		a.Steps = append(a.Steps, Step{Key: "flag:" + o.Name, Prompt: prompt, Optional: true})
	}
	a.Argv = func(answers map[string][]string) []string {
		argv := slices.Clone(path)
		for _, arg := range args {
			value := first(answers, "arg:"+arg.Name)
			switch {
			case value == "":
			case arg.Variadic:
				argv = append(argv, strings.Fields(value)...)
			default:
				argv = append(argv, value)
			}
		}
		for _, v := range answers["switches"] {
			if v != "" {
				argv = append(argv, v)
			}
		}
		for _, o := range valued {
			value := first(answers, "flag:"+o.Name)
			switch {
			case value == "":
			case o.Repeatable:
				for _, each := range strings.Fields(value) {
					argv = append(argv, "--"+o.Name, each)
				}
			default:
				argv = append(argv, "--"+o.Name, value)
			}
		}
		return argv
	}
	return a
}

func rank(b bool) int {
	if b {
		return 1
	}
	return 0
}
