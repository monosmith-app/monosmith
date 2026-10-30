package releaseassets

import (
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/spf13/cobra"
)

// Command is `smith release assets`.
func Command(version string) *cobra.Command {
	var req Request
	cmd := &cobra.Command{
		Use:   "assets",
		Short: "Build a release's files with its build command and upload them to its GitHub Release",
		Long: `assets runs, for the projects released at HEAD that have a build in
[projects.<name>.release], that command on this machine (SMITH_PLATFORM,
SMITH_VERSION and SMITH_TAG are set) and uploads the files matching their
assets globs to the GitHub Release with gh release upload --clobber. The
generated GitHub Actions workflow runs it on one runner per platform.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			req.Out = cmd.OutOrStdout()
			_, err = Run(cmd.Context(), s, req)
			return err
		},
	}
	cmd.Flags().StringVar(&req.Platform, "platform", "", "the platform of this machine: windows-x64, macos-arm64, macos-x64, linux-x64 or linux-arm64 (default: detected)")
	cmd.Flags().StringVar(&req.Project, "project", "", "build only this project")
	return cmd
}
