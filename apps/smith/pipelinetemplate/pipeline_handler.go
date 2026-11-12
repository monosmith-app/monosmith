package pipelinetemplate

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/monosmith-app/monosmith/apps/smith/commandflags"
	"github.com/monosmith-app/monosmith/apps/smith/marks"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// Command is `smith pipeline`.
func Command(version string) *cobra.Command {
	var opts Options
	var ci string
	var yes, dryRun bool
	cmd := &cobra.Command{
		Use:   "pipeline [template]",
		Short: "Add a ready-made pipeline: e2e, package-zip, deploy-pages, deploy-azure-function",
		Long: `Add a ready-made pipeline file to the workspace, beside the generated CI: it
starts the way every generated job does (the toolchains, smith, smith ci setup),
then does one thing. It is written once and yours from then on: smith upgrade
never rewrites it, and an existing file is kept unless --force. With no
template, lists them.`,
		Example: `  smith pipeline
  smith pipeline e2e
  smith pipeline deploy-pages --project web`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if len(args) == 0 {
				w := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
				for _, t := range Catalog {
					needs := ""
					if t.NeedsProject {
						needs = " (--project <app>)"
					}
					fmt.Fprintf(w, "%s%s\t%s\t%s\n", t.Name, needs, strings.Join(t.Providers, ", "), t.Description)
				}
				return w.Flush()
			}
			switch ci {
			case "":
			case "both":
				opts.Providers = []string{"github", "azure"}
			case "github", "azure":
				opts.Providers = []string{ci}
			default:
				return fmt.Errorf("--ci is github, azure or both, not %q", ci)
			}
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			lines, err := PlanLines(s, args[0], opts)
			if err != nil {
				return err
			}
			goOn, err := operationplan.Gate(out, cmd.InOrStdin(), "add the "+args[0]+" pipeline", "Pipeline files", lines, yes, dryRun)
			if err != nil || !goOn {
				return err
			}
			written, err := Add(s, args[0], opts)
			for _, f := range written {
				fmt.Fprintf(out, marks.Passed+" wrote %s\n", f)
			}
			return err
		},
	}
	commandflags.AddOneProjectFlag(cmd, &opts.Project, "the app a deploy template builds")
	cmd.Flags().StringVar(&ci, "ci", "", "github, azure or both (default: the workspace's providers)")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "replace a pipeline file that exists")
	commandflags.AddYesFlag(cmd, &yes, "do not ask; without a terminal the files are written only with --yes")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show the files that would be written and change nothing")
	return cmd
}
