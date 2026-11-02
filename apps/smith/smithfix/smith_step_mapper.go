package smithfix

import (
	"slices"

	"github.com/monosmith-app/monosmith/libs/actions/catalog"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Program is the first word of a fix step that runs smith itself. The host
// replaces it with the smith that is running, so a fix never reaches an older
// smith on the PATH.
const Program = "smith"

// Step is a fix step that runs a smith command.
//
// Parameters: description says in words what the command does; args is the
// command line without the program name.
// Returns: the step. It carries --yes when the command asks on a terminal,
// because the person already answered in the interface that shows the fix.
func Step(description string, args ...string) protocol.FixStep {
	return protocol.FixStep{
		Description: description,
		Argv:        append([]string{Program}, catalog.WithConfirmation(args)...),
	}
}

// Fix is a fix made of smith commands. It is never Safe: the person decides
// every time, and `doctor --fix --yes` alone never runs it.
//
// Parameters: description names the fix; steps run in order.
// Returns: the fix.
func Fix(description string, steps ...protocol.FixStep) *protocol.Fix {
	return &protocol.Fix{Description: description, Steps: steps}
}

// Choice is a fix that asks which smith command to run. The first option is
// the suggested one.
//
// Parameters: description names the fix; question is what the person decides;
// options are the answers, each a label and the fix it adds.
// Returns: the fix.
func Choice(description, question string, options ...protocol.FixChoiceOption) *protocol.Fix {
	return &protocol.Fix{Description: description, Choices: []protocol.FixChoice{{Question: question, Options: options}}}
}

// Option is one answer of a Choice that runs one smith command.
//
// Parameters: label is the answer as the person reads it, with what it does;
// args is the command line without the program name.
// Returns: the option.
func Option(label string, args ...string) protocol.FixChoiceOption {
	return protocol.FixChoiceOption{Label: label, Fix: protocol.Fix{Description: label, Steps: []protocol.FixStep{Step(label, args...)}}}
}

// IsStep reports whether a step's command line runs smith.
func IsStep(argv []string) bool { return len(argv) > 1 && argv[0] == Program }

// PlanArgv is the command line that prints what a smith step would do and
// changes nothing, for the review of a fix.
//
// Parameters: argv is the step's command line.
// Returns: the plan command line, and false when the step does not run smith
// or the command has no plan flag.
func PlanArgv(argv []string) ([]string, bool) {
	if !IsStep(argv) {
		return nil, false
	}
	args := slices.DeleteFunc(slices.Clone(argv[1:]), func(a string) bool { return a == "--yes" || a == "-y" })
	flag, ok := catalog.PlanFlag(args)
	if !ok {
		return nil, false
	}
	return append(append([]string{Program}, args...), flag), true
}
