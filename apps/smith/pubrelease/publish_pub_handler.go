package pubrelease

import (
	"errors"
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/commandflags"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// Command is `smith release pub`.
func Command(version string) *cobra.Command {
	var req Request
	cmd := &cobra.Command{
		Use:   "pub",
		Short: "Publish a released app to pub.dev from its GitHub Release (the generated pub workflow runs it)",
		Long: `pub builds an app's pub.dev package from the archives on its GitHub Release
and publishes it with dart pub publish, unless pub.dev already has the version.
The release starts the generated .github/workflows/pub.yml on the app's pub
tag, which runs this: pub.dev's automated publishing accepts only such a run.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if req.Project == "" || req.Version == "" {
				return errors.New("--project and --version are required")
			}
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			res, err := Publish(s, req)
			out := cmd.OutOrStdout()
			if len(res.Skipped) > 0 {
				fmt.Fprintf(out, "no Dart build for %v\n", res.Skipped)
			}
			switch {
			case err != nil:
				return err
			case res.Existing != "":
				fmt.Fprintf(out, "pub.dev already has %s\n", res.Existing)
			default:
				fmt.Fprintf(out, "published %s to pub.dev\n", res.Published)
			}
			return nil
		},
	}
	commandflags.AddOneProjectFlag(cmd, &req.Project, "the released app")
	cmd.Flags().StringVar(&req.Version, "version", "", "its released version")
	return cmd
}
