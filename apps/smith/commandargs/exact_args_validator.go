package commandargs

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/spf13/cobra"
)

// Spec says what a command's positional arguments are.
type Spec struct {
	// Names describe each argument in order, as the person would say it:
	// "a kind", "a name for the project".
	Names []string
	// Example is a complete command line that works.
	Example string
	// More is an extra remedy, such as "smith kinds lists the kinds".
	More string
}

// Exactly accepts exactly len(spec.Names) arguments. A wrong count becomes a
// problem naming what is missing (or what is extra) with an example.
func Exactly(spec Spec) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		n := len(spec.Names)
		if len(args) == n {
			return nil
		}
		var p string
		if len(args) < n {
			p = cmd.CommandPath() + " needs " + joinNames(spec.Names[len(args):])
		} else {
			p = cmd.CommandPath() + " takes " + joinNames(spec.Names) + " and nothing else; got \"" + strings.Join(args[n:], " ") + "\" too"
		}
		remedy := []string{"for example: " + spec.Example}
		if spec.More != "" {
			remedy = append(remedy, spec.More)
		}
		return problem.New(p, remedy...)
	}
}

func joinNames(names []string) string {
	switch len(names) {
	case 0:
		return "no arguments"
	case 1:
		return names[0]
	}
	return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
}

// NoneBut accepts no arguments. When one is given it is a mistake the person
// is likely to make, so redirect(arg) says what they probably meant.
func NoneBut(redirect func(arg string) (problemText, remedy string)) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}
		p, r := redirect(args[0])
		return problem.New(p, r)
	}
}
