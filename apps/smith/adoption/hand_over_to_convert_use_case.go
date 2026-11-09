package adoption

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// Convertible is true when the folder is an Nx or Turborepo workspace that is
// not a smith workspace yet: smith convert's, which adopt hands it to.
func Convertible(dir string) bool {
	exists := func(rel string) bool {
		_, err := os.Stat(filepath.Join(dir, rel))
		return err == nil
	}
	return !exists(workspace.FileName) && (exists("nx.json") || exists("turbo.json"))
}

// handOverToConvert runs smith convert for an Nx or Turborepo workspace, so
// that adopt is the one way in: a report is convert's --dry-run and --init is
// convert applying.
func handOverToConvert(cmd *cobra.Command, level string, asJSON, dryRun, initialise, tags, install, eslint, commit bool) error {
	switch {
	case tags:
		return problem.New("--tags does not apply to an Nx or Turborepo workspace: the conversion carries its release tags over", "smith adopt --init")
	case install || eslint:
		return problem.New("--install and --eslint do not apply to an Nx or Turborepo workspace: the conversion installs, and asks about lint, itself", "smith adopt --init")
	case dryRun && initialise:
		return problem.New("--dry-run reports and changes nothing, so it cannot be combined with --init", "drop --dry-run to apply, or drop --init to only look")
	}
	conv, _, err := cmd.Root().Find([]string{"convert"})
	if err != nil || conv == nil || conv.Name() != "convert" || conv.RunE == nil {
		return problem.New("this is an Nx or Turborepo workspace, which smith convert brings in", "smith convert --dry-run")
	}
	if !asJSON {
		fmt.Fprintln(cmd.ErrOrStderr(), "This is an Nx or Turborepo workspace, so smith converts it (the same as smith convert).")
	}
	flags := map[string]string{"level": level}
	if !initialise {
		flags["dry-run"] = "true"
	}
	if asJSON {
		flags["json"] = "true"
	}
	if commit {
		flags["commit"] = "true"
	}
	for name, value := range flags {
		if err := conv.Flags().Set(name, value); err != nil {
			return err
		}
	}
	conv.SetOut(cmd.OutOrStdout())
	conv.SetErr(cmd.ErrOrStderr())
	return conv.RunE(conv, nil)
}
