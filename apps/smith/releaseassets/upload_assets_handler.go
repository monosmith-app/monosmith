package releaseassets

import (
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
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
generated GitHub Actions workflow runs it on one runner per platform.

With --tag project@version it runs again for a release that exists. When the
tag is not at HEAD, --source says which code to build. current is the code as
it is now, named with the tag's version: a fix to the build or to smith.toml
reaches the files, but they carry newer code under an old number. tag is the
tag's own commit, built in a temporary git worktree: the exact code that was
released, with the smith.toml of that commit. Without --source a terminal is
asked and anything else is refused. Files of the same name are replaced.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := session.Open(session.Options{Dir: ".", Version: version, Log: cmd.ErrOrStderr()})
			if err != nil {
				return err
			}
			defer s.Close()
			req.Out, req.In, req.Terminal = cmd.OutOrStdout(), cmd.InOrStdin(), operationplan.DetectEnvironment().Terminal
			_, err = Run(cmd.Context(), s, req)
			return err
		},
	}
	cmd.Flags().StringVar(&req.Platform, "platform", "", "the platform of this machine: windows-x64, macos-arm64, macos-x64, linux-x64 or linux-arm64 (default: detected)")
	cmd.Flags().StringVar(&req.Project, "project", "", "build only this project")
	cmd.Flags().StringVar(&req.Tag, "tag", "", "build for an existing release tag (project@version), replacing its files")
	cmd.Flags().StringVar(&req.Source, "source", "", "with --tag, the code to build when the tag is not at HEAD: current (the code as it is now, named with the tag's version) or tag (the tag's own commit); default: ask in a terminal, refuse without one")
	return cmd
}
