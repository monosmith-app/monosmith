package convertchoices

import "github.com/monosmith-app/monosmith/apps/smith/legacyinventory"

// Questions are the choices as questions, each with its suggested option
// (the one in def) first.
func Questions(u legacyinventory.Usage, def Choices) []Question {
	lint := []Option{
		{Smith, "smith's ESLint config, which also keeps code organised by feature; the old lint packages and config files are removed"},
		{Keep, "keep the current lint setup"},
	}
	if !u.ESLint {
		lint = []Option{{Keep, "keep (this workspace has no ESLint config to move)"}}
	}
	test := []Option{
		{Jest, "Jest; the configs and packages of other runners are removed"},
		{Vitest, "Vitest; the configs and packages of other runners are removed"},
		{Keep, "keep the current test setup"},
	}
	build := []Option{
		{Smith, "smith's build for the stack; the old build configs and packages are removed"},
		{Keep, "keep the current build setup"},
	}
	ci := []Option{
		{Smith, "the pipeline smith generates"},
		{Keep, "keep the current pipeline (it keeps calling the old tool until you change it)"},
	}
	return []Question{
		{Key: "lint", Flag: "--lint", Question: "Lint with:", Options: first(lint, def.Lint)},
		{Key: "test", Flag: "--test", Question: "Run tests with:", Options: first(test, def.Test)},
		{Key: "build", Flag: "--build", Question: "Build with:", Options: first(build, def.Build)},
		{Key: "ci", Flag: "--ci", Question: "CI pipeline:", Options: first(ci, def.CI)},
	}
}

// first moves the option with the value to the front, keeping the others'
// order.
func first(options []Option, value string) []Option {
	out := make([]Option, 0, len(options))
	for _, o := range options {
		if o.Value == value {
			out = append(out, o)
		}
	}
	for _, o := range options {
		if o.Value != value {
			out = append(out, o)
		}
	}
	return out
}
