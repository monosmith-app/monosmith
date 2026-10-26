package slicecheck

import (
	"fmt"
	"strings"

	archcheck "github.com/monosmith-app/monosmith/libs/plugin-architectures/check"
	"github.com/monosmith-app/monosmith/libs/plugin-architectures/imports"
	"github.com/monosmith-app/monosmith/libs/plugin-architectures/rules"
	"github.com/monosmith-app/monosmith/libs/plugin-vertical-slices/languages"
	"github.com/monosmith-app/monosmith/libs/plugin-vertical-slices/layout"
	"github.com/spf13/cobra"
)

// Command is `smith slices`.
func Command() *cobra.Command {
	cmd := &cobra.Command{Use: "slices", Short: "Check the vertical-slice layout of projects"}
	var lang, architecture string
	check := &cobra.Command{
		Use:   "check <project directory>...",
		Short: "Check that files are role-named and no folder is a junk drawer",
		Long: `Checks Go, Python, Dart, Rust, Ruby, C++, .NET, JVM or PHP projects against the vertical-slice rules: a file is
<name>_<role> (roles: ` + fmt.Sprint(languages.Roles) + `),
a module's or package's root holds only its entry files, and no folder is
named util, helper, common, shared, manager, processor, data or misc. A slice
with more than ` + fmt.Sprint(layout.SliceSize) + ` production files is reported for review.

The vertical-slices convention ([conventions.vertical-slices] in smith.toml)
runs this as each project's slice-check task.

With --architecture the check is another architecture's instead: it reads
the imports of the project's sources and reports the ones that architecture
forbids. The convention of that architecture (for example
[conventions.hexagonal]) runs it as each project's architecture-check task.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, dirs []string) error {
			if architecture != "" {
				return checkArchitecture(cmd, architecture, lang, dirs)
			}
			problems := 0
			for _, dir := range dirs {
				r, err := layout.Check(dir, lang)
				if err != nil {
					return err
				}
				for _, w := range r.Warnings {
					fmt.Fprintln(cmd.OutOrStdout(), "review: "+w)
				}
				for _, p := range r.Problems {
					fmt.Fprintln(cmd.ErrOrStderr(), p)
				}
				problems += len(r.Problems)
			}
			if problems > 0 {
				return fmt.Errorf("%d file(s) or folder(s) break the vertical-slice layout; an exception is written in the project README", problems)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "vertical-slice layout OK")
			return nil
		},
	}
	check.Flags().StringVar(&lang, "lang", "go", "language of the projects: "+strings.Join(languages.Names(), ", ")+"; with --architecture also node")
	check.Flags().StringVar(&architecture, "architecture", "", "check this architecture's import rules instead: "+strings.Join(rules.Names(), ", "))
	cmd.AddCommand(check)
	return cmd
}

// checkArchitecture reports the imports of the projects in dirs that the
// architecture forbids.
func checkArchitecture(cmd *cobra.Command, name, lang string, dirs []string) error {
	rs, ok := rules.For(name)
	if !ok {
		return fmt.Errorf("no architecture called %q with import rules; the choices are %s", name, strings.Join(rules.Names(), ", "))
	}
	if _, ok := imports.For(lang); !ok {
		return fmt.Errorf("the architecture check knows %s, not %q", strings.Join(imports.Names(), ", "), lang)
	}
	problems := 0
	for _, dir := range dirs {
		r, err := archcheck.Check(dir, lang, rs)
		if err != nil {
			return err
		}
		for _, w := range r.Warnings {
			fmt.Fprintln(cmd.OutOrStdout(), "review: "+w)
		}
		for _, p := range r.Problems {
			fmt.Fprintln(cmd.ErrOrStderr(), p)
		}
		problems += len(r.Problems)
	}
	if problems > 0 {
		return fmt.Errorf("%d import(s) break the %s architecture; an exception is written in the project README", problems, name)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s architecture OK\n", name)
	return nil
}
