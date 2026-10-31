package report

import (
	"fmt"
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// Command is `smith report`.
func Command(version string) *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Write one HTML page: the workspace map, health, derived-file drift, choices and release plan",
		Long: `Write one self-contained HTML page about the workspace, to attach to a pull
request or to open in a browser.

The page shows the workspace map (an inline dependency graph, with the same
facts as a table), the doctor results, the derived files that need action, the
stack and choices of each project, and the release plan: what a merge of the
current commit to the release branch would release.

The page has no script and makes no request. Every status is a word, not only a
colour. The page follows the light or dark setting of the browser. The release
plan reads the commits since each package's last tag and asks no registry, so it
is a preview, not a rehearsal: smith release --dry-run checks more.

Without -o the page goes to the standard output.`,
		Example: "  smith report -o smith-report.html\n  smith report > report.html",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr(), AllowCycles: true})
			if err != nil {
				return err
			}
			defer s.Close()
			out, err := Render(Gather(s))
			if err != nil {
				return err
			}
			if file == "" {
				fmt.Fprint(cmd.OutOrStdout(), out)
				return nil
			}
			if err := os.WriteFile(file, []byte(out), 0o644); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "wrote "+file)
			return nil
		},
	}
	cmd.Flags().StringVarP(&file, "output", "o", "", "write the page to a file")
	return cmd
}
